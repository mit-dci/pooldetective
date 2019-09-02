package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/mit-dci/pooldetective/logging"
)

var dockerClient *client.Client
var networkID string

var (
	ErrContainerNotFound = errors.New("Container not found")
	ErrImageNotFound     = errors.New("Image not found")
)

func getPrefixedImageName(image string) string {
	img := ""
	registry := os.Getenv("REGISTRY")
	prefix := os.Getenv("IMAGE_PREFIX")
	if prefix == "" {
		prefix = "pooldetective"
	}
	if registry != "" {
		img += registry + "/"
	}
	img += fmt.Sprintf("%s-%s", prefix, image)
	return img
}

func initDocker() {
	var err error
	dockerClient, err = client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}
	client.FromEnv(dockerClient)
}

func initNetwork() error {
	networks, err := dockerClient.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return err
	}

	networkName := os.Getenv("DOCKER_NETWORK")
	if networkName == "" {
		networkName = "pooldetective"
	}
	found := false
	for _, n := range networks {
		if n.Name == networkName {
			found = true
			networkID = n.ID
			logging.Infof("Network [%s] exists", networkName)
			break

		}
	}

	if !found {
		logging.Infof("Network [%s] does not exist, creating...", networkName)
		res, err := dockerClient.NetworkCreate(context.Background(), networkName, types.NetworkCreate{})
		if err != nil {
			return err
		}
		networkID = res.ID
		logging.Infof("Network [%s] created", networkName)
	}
	return nil
}

func containerName(component string, id int) string {
	containerNamePrefix := os.Getenv("CONTAINER_NAME_PREFIX")
	if containerNamePrefix == "" {
		containerNamePrefix = "pooldetective"
	}
	return fmt.Sprintf("%s-%s-%d", containerNamePrefix, component, id)
}

func getImage(name string) (string, error) {
	res, err := dockerClient.ImageList(context.Background(), types.ImageListOptions{All: true})
	if err != nil {
		return "", err
	}

	for _, r := range res {
		for _, tag := range r.RepoTags {
			if tag == fmt.Sprintf("%s:latest", name) {
				return r.ID, nil
			}
		}
	}
	return "", ErrImageNotFound
}

func pullLatestImages() error {
	if os.Getenv("REGISTRY") != "" {
		images := []string{
			getPrefixedImageName("stratumclient"),
			getPrefixedImageName("stratumserver"),
			getPrefixedImageName("blockobserver"),
		}
		for _, i := range images {
			io, err := dockerClient.ImagePull(context.Background(), fmt.Sprintf("%s:latest", i), types.ImagePullOptions{
				RegistryAuth: os.Getenv("REGISTRYAUTH"),
			})
			if err != nil {
				return err
			}
			result, err := ioutil.ReadAll(io)
			if err != nil {
				return err
			}
			if !strings.Contains(string(result), "Image is up to date") {
				logging.Debugf("Fetch result for image %s: %s", i, string(result))
			}
			io.Close()
		}
	}
	return nil
}

func checkBlockObserver() error {
	logging.Infof("Checking block observer")

	// Get image we need to use
	imageID, err := getImage(getPrefixedImageName("blockobserver"))
	if err != nil {
		return err
	}

	name := containerName("blockobserver", 0)
	name = strings.ReplaceAll(name, "-0", "")
	cnt, err := getContainer(name)
	if err != nil {
		if err != ErrContainerNotFound {
			return err
		}
		// Create container
		logging.Info("Block observer does not exist, creating...")
		containerConfig := new(container.Config)
		containerConfig.Image = imageID
		containerConfig.Env = []string{
			fmt.Sprintf("LOCATIONID=%d", locationID),
			"LOGLEVEL=4",
			fmt.Sprintf("HUBHOST=%s", os.Getenv("HUBHOST")),
		}

		c, err := dockerClient.ContainerCreate(context.Background(), containerConfig, nil, nil, name)
		if err != nil {
			return err
		}

		logging.Info("Connecting block observer to network...")
		err = dockerClient.NetworkConnect(context.Background(), networkID, c.ID, nil)
		if err != nil {
			return err
		}

		logging.Info("Block observer  created!")
		err = dockerClient.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})
		time.Sleep(time.Second * 5)
		return err
	}

	if cnt.ImageID != imageID {
		// Needs upgrade
		logging.Warn("Image of block observer is outdated, removing...")
		err := stopAndRemoveContainer(cnt.ID)
		if err != nil {
			return err
		}
		return checkBlockObserver()
	}

	if cnt.State != "running" {
		logging.Debug("Container for block observer is not running, starting..")
		err = dockerClient.ContainerStart(context.Background(), cnt.ID, types.ContainerStartOptions{})
		time.Sleep(time.Second * 5)
		return err
	}

	logging.Debugf("Block observer is OK!")

	return nil
}

func restartStratumClient(poolObserverID int) error {
	name := containerName("stratum-client", poolObserverID)
	cnt, err := getContainer(name)
	if err != nil {
		return err
	}

	return restartContainer(cnt.ID)
}

func checkStratumClient(poolObserverID int) error {
	logging.Infof("Checking stratum client %d", poolObserverID)

	// Get image we need to use
	imageID, err := getImage(getPrefixedImageName("stratumclient"))
	if err != nil {
		return err
	}

	name := containerName("stratum-client", poolObserverID)
	cnt, err := getContainer(name)
	if err != nil {
		if err != ErrContainerNotFound {
			return err
		}
		// Create container
		logging.Infof("Stratum client %d does not exist, creating...", poolObserverID)
		containerConfig := new(container.Config)
		containerConfig.Image = imageID
		containerConfig.Env = []string{
			fmt.Sprintf("POOLOBSERVERID=%d", poolObserverID),
			"LOGLEVEL=4",
			fmt.Sprintf("HUBHOST=%s", os.Getenv("HUBHOST")),
		}

		c, err := dockerClient.ContainerCreate(context.Background(), containerConfig, nil, nil, name)
		if err != nil {
			return err
		}

		logging.Infof("Connecting stratum client %d to network...", poolObserverID)
		err = dockerClient.NetworkConnect(context.Background(), networkID, c.ID, nil)
		if err != nil {
			return err
		}

		logging.Infof("Stratum client %d created!", poolObserverID)
		err = dockerClient.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})
		time.Sleep(time.Second * 5)
		return err
	}

	if cnt.ImageID != imageID {
		// Needs upgrade
		logging.Warnf("Image of stratum client %d is outdated, removing...", poolObserverID)
		err := stopAndRemoveContainer(cnt.ID)
		if err != nil {
			return err
		}
		return checkStratumClient(poolObserverID)
	}

	if cnt.State != "running" {
		logging.Debugf("Container for stratum client %d is not running, starting..", poolObserverID)
		err = dockerClient.ContainerStart(context.Background(), cnt.ID, types.ContainerStartOptions{})
		time.Sleep(time.Second * 5)
		return err
	}

	logging.Debugf("Stratum client %d is OK!", poolObserverID)

	return nil
}

func removeUnneededContainers(prefix string, IDs []int) error {
	prefix = strings.ReplaceAll(containerName(prefix, 0), "-0", "-")
	containers, err := getContainersWithPrefix(prefix)
	if err != nil {
		return err
	}
	for _, c := range containers {
		match := false
		for _, n := range c.Names {
			for _, a := range IDs {
				name := fmt.Sprintf("%s%d", prefix, a)
				if n[1:] == name {
					match = true
					break
				}
			}
			if match == true {
				break
			}
		}
		if !match {
			logging.Debugf("Removing container [%s] because it's no longer in the config")
			err = stopAndRemoveContainer(c.ID)
			time.Sleep(time.Second * 5)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkStratumServer(algorithmID, port, stratumProtocol int) error {
	logging.Infof("Checking stratum server %d", algorithmID)

	// Get image we need to use
	imageID, err := getImage(getPrefixedImageName("stratumserver"))
	if err != nil {
		return err
	}

	logging.Infof("Image ID is %s", imageID)

	name := containerName("stratum-server", algorithmID)
	cnt, err := getContainer(name)
	if err != nil {
		if err != ErrContainerNotFound {
			return err
		}
		// Create container
		logging.Infof("Stratum server %d does not exist, creating...", algorithmID)
		containerConfig := new(container.Config)
		containerConfig.Image = imageID
		containerConfig.Env = []string{
			fmt.Sprintf("STRATUMPROTOCOL=%d", stratumProtocol),
			fmt.Sprintf("ALGORITHMID=%d", algorithmID),
			fmt.Sprintf("STRATUMPORT=%d", port),
			"LOGLEVEL=4",
			fmt.Sprintf("HUBHOST=%s", os.Getenv("HUBHOST")),
			fmt.Sprintf("PGSQL_CONNECTION=%s", os.Getenv("PGSQL_CONNECTION")),
		}
		containerConfig.ExposedPorts = nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", port)): struct{}{},
		}

		hostConfig := new(container.HostConfig)

		hostConfig.PortBindings = nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", port)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprintf("%d", port),
				},
			},
		}

		c, err := dockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, name)
		if err != nil {
			return err
		}

		logging.Infof("Connecting stratum server %d to network...", algorithmID)
		err = dockerClient.NetworkConnect(context.Background(), networkID, c.ID, nil)
		if err != nil {
			return err
		}

		logging.Infof("Stratum server %d created!", algorithmID)
		err = dockerClient.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})
		time.Sleep(time.Second * 5)
		return err
	}
	if cnt.ImageID != imageID {
		// Needs upgrade
		logging.Warnf("Image of stratum server %d is outdated, removing...", algorithmID)
		err := stopAndRemoveContainer(cnt.ID)
		if err != nil {
			return fmt.Errorf("Could not stop and remove container: %v", err)
		}
		return checkStratumServer(algorithmID, port, stratumProtocol)
	}
	if cnt.State != "running" {
		logging.Debugf("Container for stratum server %d is not running, starting..", algorithmID)
		err = dockerClient.ContainerStart(context.Background(), cnt.ID, types.ContainerStartOptions{})
		time.Sleep(time.Second * 5)
		return err
	}
	return nil
}

func getContainersWithPrefix(name string) ([]types.Container, error) {
	result := []types.Container{}
	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return result, err
	}

	for _, c := range containers {
		for _, n := range c.Names {
			if strings.HasPrefix(n[1:], name) {
				result = append(result, c)
				break
			}
		}
	}
	return result, nil
}

func getContainer(name string) (types.Container, error) {
	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return types.Container{}, err
	}

	for _, c := range containers {
		for _, n := range c.Names {

			if n[1:] == name {
				return c, nil
			}
		}
	}
	return types.Container{}, ErrContainerNotFound
}

func removeContainer(containerID string) error {
	logging.Infof("Removing container %s", containerID)
	opt := types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}
	return dockerClient.ContainerRemove(context.Background(), containerID, opt)
}

func stopContainer(containerID string) error {
	logging.Infof("Stopping container %s", containerID)
	timeout := time.Duration(time.Minute)
	return dockerClient.ContainerStop(context.Background(), containerID, &timeout)
}

func startContainer(containerID string) error {
	logging.Infof("Starting container %s", containerID)
	return dockerClient.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
}

func stopAndRemoveContainer(containerID string) error {
	err := stopContainer(containerID)
	if err != nil {
		return err
	}
	return removeContainer(containerID)
}

func restartContainer(containerID string) error {
	err := stopContainer(containerID)
	if err != nil {
		return err
	}

	return startContainer(containerID)
}
