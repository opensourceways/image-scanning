package domain

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opensourceways/server-common-lib/utils"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/image-scanning/scanning/domain/primitive"
	localutils "github.com/opensourceways/image-scanning/utils"
)

const (
	registryDocker = "docker.io"
	registryQuay   = "quay.io"

	// apiToListTagsOfDocker reference: https://docs.docker.com/reference/api/hub/latest/#tag/repositories/operation/ListRepositoryTags
	apiToListTagsOfDocker = "https://hub.docker.com/v2/namespaces/%s/repositories/%s/tags?page_size=100"
	//apiToListTagsOfQuay reference: https://docs.redhat.com/en/documentation/red_hat_quay/3.6/html-single/red_hat_quay_api_guide/index#listrepotags
	apiToListTagsOfQuay = "https://quay.io//api/v1/repository/%s/%s/tag/?limit=100&page=%d"
)

var (
	globalConfig map[string]*Global
)

func initGlobalConfig(community string, cfg *Global) {
	if len(globalConfig) == 0 {
		globalConfig = make(map[string]*Global)
	}

	globalConfig[community] = cfg
}

type ScanConfig struct {
	Version string  `json:"version"`
	Scanner Scanner `json:"scanner"`
	Repos   []Repo  `json:"repos"`
	Images  []Image `json:"images"`
}

type Scanner struct {
	Global Global `json:"global"`
}

type Global struct {
	DefaultArches   []string `json:"default_arches"`
	DefaultInterval string   `json:"default_interval"`

	Output Output `json:"output"`
}

type Output struct {
	Repo string `json:"repo"`
	Path string `json:"path"`
}

type Repo struct {
	Namespace string   `json:"namespace"`
	Registry  string   `json:"registry"`
	Images    []string `json:"images"`
	Arches    []string `json:"arches"`
	Interval  string   `json:"interval"`
}

type Image struct {
	Image string `json:"image"`
	Tags  []Tag  `json:"tags"`
}

type Tag struct {
	Tag      string   `json:"tag"`
	Interval string   `json:"interval"`
	Arches   []string `json:"arches"`
	Disable  bool     `json:"disable"`
}

func (r Repo) genTask(communityName string, tasks map[string]Task) {
	arch := getArches(communityName, r.Arches)
	interval, err := getInterval(communityName, r.Interval)
	if err != nil {
		logrus.Errorf("get interval of %s failed: %s", r.Namespace, err.Error())
		return
	}

	for _, image := range r.Images {
		tags, err := r.AllTagsOfImage(image)
		if err != nil {
			logrus.Errorf("get all tags of %s/%s failed: %s", r.Namespace, image, err.Error())
			continue
		}

		for _, tag := range tags {
			task, err := ToTask(communityName, r.Registry, r.Namespace, image, tag, arch, interval)
			if err != nil {
				logrus.Errorf("repo to task failed: %s", err.Error())
				continue
			}

			tasks[task.UniqueKey()] = task
		}
	}
}

func (i Image) genTask(communityName string, tasks map[string]Task) {
	for _, tag := range i.Tags {
		task, err := tag.ToTask(communityName)
		if err != nil {
			logrus.Errorf("tag %s to task failed: %s", tag.Tag, err.Error())
			continue
		}

		// 当批量扫描任务和精准扫描任务发生重叠时，以精准的任务为高优先级，覆盖批量数据
		tasks[task.UniqueKey()] = task
	}
}

func (t Tag) ToTask(communityName string) (task Task, err error) {
	if t.Disable {
		err = errors.New("tag is disabled")
		return
	}

	arch := getArches(communityName, t.Arches)
	interval, err := getInterval(communityName, t.Interval)
	if err != nil {
		return
	}

	split := strings.Split(t.Tag, "/")
	if len(split) != 3 {
		err = errors.New("tag format error")
		return
	}

	registry := split[0]
	namespace := split[1]
	imageAndTag := split[2]
	split2 := strings.Split(imageAndTag, ":")
	if len(split2) != 2 {
		err = errors.New("image:tag format error")
		return
	}

	return ToTask(communityName, registry, namespace, split2[0], split2[1], arch, interval)
}

func (r Repo) AllTagsOfImage(image string) ([]string, error) {
	switch r.Registry {
	case registryDocker:
		return r.getTagsFromDocker(image)
	case registryQuay:
		return r.getTagsFromQuay(image)
	default:
		return nil, errors.New("unsupported registry")
	}
}

type tagsResponseOfDocker struct {
	Count   int    `json:"count"`
	Next    string `json:"next"`
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func (r Repo) getTagsFromDocker(image string) ([]string, error) {
	url := fmt.Sprintf(apiToListTagsOfDocker, r.Namespace, image)
	client := utils.NewHttpClient(3)

	var tags []string
	for {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var resp tagsResponseOfDocker
		if _, err = client.ForwardTo(req, &resp); err != nil {
			return nil, err
		}

		for _, v := range resp.Results {
			tags = append(tags, v.Name)
		}

		if resp.Next == "" {
			break
		} else {
			url = resp.Next
		}

		// docker的api每分钟限速180次
		// 这里设置大概每分钟访问120次
		time.Sleep(time.Millisecond * 500)
	}

	return tags, nil
}

type tagsResponseOfQuay struct {
	Page          int  `json:"page"`
	HasAdditional bool `json:"has_additional"`
	Tags          []struct {
		Name string `json:"name"`
	} `json:"tags"`
}

func (r Repo) getTagsFromQuay(image string) ([]string, error) {
	page := 1
	client := utils.NewHttpClient(3)
	var tags []string
	for {
		url := fmt.Sprintf(apiToListTagsOfQuay, r.Namespace, image, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var resp tagsResponseOfQuay
		if _, err = client.ForwardTo(req, &resp); err != nil {
			return nil, err
		}

		for _, v := range resp.Tags {
			tags = append(tags, v.Name)
		}

		if !resp.HasAdditional {
			break
		} else {
			page++
		}
	}

	return tags, nil
}

func getArches(community string, arches []string) []string {
	if len(arches) != 0 {
		return arches
	}

	global, ok := globalConfig[community]
	if !ok {
		return nil
	}

	return global.DefaultArches
}

func getInterval(community, intervalString string) (int, error) {
	interval, err := localutils.StringToInterval(intervalString)
	if err == nil {
		return interval, nil
	}

	global, ok := globalConfig[community]
	if !ok {
		return 0, errors.New("no interval")
	}

	return localutils.StringToInterval(global.DefaultInterval)
}

func ToTask(communityName, registry, namespace, image, tag string, arch []string, interval int) (Task, error) {
	pRegistry, err := primitive.NewRegistry(registry)
	if err != nil {
		return Task{}, err
	}

	return Task{
		Community: communityName,
		Registry:  pRegistry,
		Namespace: namespace,
		Image:     image,
		Tag:       tag,
		Arch:      arch,
		Interval:  interval,
	}, nil
}
