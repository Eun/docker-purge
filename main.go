package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Eun/docker-purge/jq"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"gopkg.in/alecthomas/kingpin.v2"
)

//go:generate go run generate.go

// VersionHash represents the git sha1 which this version was built on
var VersionHash = "Unknown/CustomBuild"

// Version represents the build version of docker-purge
var Version = "Unknown/CustomBuild"

// BuildDate represents the build date of docker-purge
var BuildDate = "Unknown/CustomBuild"

var (
	filterArg = kingpin.Arg("filter", "jq filter to apply").String()
	// list
	listAllFlag       = kingpin.Flag("list-all", "list docker containers, images, networks").Bool()
	listContainerFlag = kingpin.Flag("list-containers", "list docker containers").Bool()
	listImageFlag     = kingpin.Flag("list-images", "list docker images").Bool()
	listNetworkFlag   = kingpin.Flag("list-networks", "list docker networks").Bool()

	dryRunFlag = kingpin.Flag("dry", "dry run, do not purge anything").Short('d').Bool()

	// limit
	limitToContainerFlag = kingpin.Flag("containers", "limit purge to docker containers").Bool()
	limitToImageFlag     = kingpin.Flag("images", "limit purge to docker images").Bool()
	limitToNetworkFlag   = kingpin.Flag("networks", "limit purge to docker networks").Bool()

	forceRemoveFlag = kingpin.Flag("force", "sets container.remove.force and image.remove.force to true").Bool()
	removeAllFlag   = kingpin.Flag("all", "remove everything related to an entity").Bool()

	// container remove options
	containerRemoveForceFlag   = kingpin.Flag("container.remove.force", "force removal of container").Bool()
	containerRemoveLinksFlag   = kingpin.Flag("container.remove.links", "remove links during removal").Bool()
	containerRemoveVolumesFlag = kingpin.Flag("container.remove.volumes", "remove volumes during removal").Bool()
	containerStop              = kingpin.Flag("container.stop", "stop running docker container").Bool()
	containerKillSignal        = kingpin.Flag("container.kill", "kill running docker container with the specified signal").Default("").String()

	// image remove options
	imageRemoveForceFlag         = kingpin.Flag("image.remove.force", "force removal of image").Bool()
	imageRemovePruneChildrenFlag = kingpin.Flag("image.remove.prunechildren", "prune children on removal").Bool()
)

var containerListOptions = types.ContainerListOptions{
	All: true,
}

var imageListOptions = types.ImageListOptions{
	All: true,
}

var networkListOptions = types.NetworkListOptions{}

var containerRemoveOptions types.ContainerRemoveOptions
var imageRemoveOptions types.ImageRemoveOptions

type container struct {
	IsImage     bool
	IsContainer bool
	IsNetwork   bool
	types.Container
}

type image struct {
	IsImage     bool
	IsContainer bool
	IsNetwork   bool
	types.ImageSummary
}

type network struct {
	IsImage     bool
	IsContainer bool
	IsNetwork   bool
	types.NetworkResource
}

func main() {
	kingpin.Parse()

	if !jq.IsValidFilter(*filterArg) {
		fmt.Fprintf(os.Stderr, "Invalid filter `%s'\n", *filterArg)
		os.Exit(1)
	}

	if *forceRemoveFlag {
		*containerRemoveForceFlag = true
		*imageRemoveForceFlag = true
		*containerStop = true
		*containerKillSignal = "9"
	}

	if *removeAllFlag {
		*containerRemoveLinksFlag = true
		*containerRemoveVolumesFlag = true
		*imageRemovePruneChildrenFlag = true
	}

	containerRemoveOptions = types.ContainerRemoveOptions{
		Force:         *containerRemoveForceFlag,
		RemoveLinks:   *containerRemoveLinksFlag,
		RemoveVolumes: *containerRemoveVolumesFlag,
	}

	imageRemoveOptions = types.ImageRemoveOptions{
		Force:         *imageRemoveForceFlag,
		PruneChildren: *imageRemovePruneChildrenFlag,
	}

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer dockerClient.Close()

	handleListFlags(dockerClient)
	handlePurge(dockerClient)

	os.Exit(0)
}

func handleListFlags(dockerClient *client.Client) {
	if *listAllFlag {
		*listContainerFlag = true
		*listImageFlag = true
		*listNetworkFlag = true
	}

	var allEntities []interface{}

	if *listContainerFlag {
		entities, err := selectContainers(dockerClient, *filterArg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		allEntities = append(allEntities, entities)
	}

	if *listImageFlag {
		entities, err := selectImages(dockerClient, *filterArg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		allEntities = append(allEntities, entities)
	}

	if *listNetworkFlag {
		entities, err := selectNetworks(dockerClient, *filterArg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		allEntities = append(allEntities, entities)
	}

	if *listContainerFlag || *listImageFlag || *listNetworkFlag {
		if len(allEntities) == 0 {
			fmt.Println("[]")
		} else {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(allEntities)
		}
		os.Exit(0)
	}
}

func handlePurge(dockerClient *client.Client) {
	if *dryRunFlag {
		fmt.Fprintln(os.Stdout, "Dry mode on")
	}

	if !*limitToContainerFlag && !*limitToImageFlag && !*limitToNetworkFlag {
		*limitToContainerFlag = true
		*limitToImageFlag = true
		*limitToNetworkFlag = true
	}

	if *limitToContainerFlag {
		containersToDelete, err := selectContainers(dockerClient, *filterArg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		if !*dryRunFlag {
			deleteContainers(dockerClient, containersToDelete, containerRemoveOptions, *containerKillSignal, *containerStop)
		} else {
			for _, container := range containersToDelete {
				fmt.Fprintf(os.Stdout, "Would delete container %s\n", container.ID)
			}
		}
	}

	if *limitToImageFlag {
		imagesToDelete, err := selectImages(dockerClient, *filterArg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		if !*dryRunFlag {
			deleteImages(dockerClient, imagesToDelete, imageRemoveOptions)
		} else {
			for _, image := range imagesToDelete {
				fmt.Fprintf(os.Stdout, "Would delete image %s\n", image.ID)
			}
		}
	}

	if *limitToNetworkFlag {
		networksToDelete, err := selectNetworks(dockerClient, *filterArg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		if !*dryRunFlag {
			deleteNetworks(dockerClient, networksToDelete)
		} else {
			for _, image := range networksToDelete {
				fmt.Fprintf(os.Stdout, "Would delete image %s\n", image.ID)
			}
		}
	}
}

func selectContainers(dockerClient *client.Client, filter string) ([]container, error) {
	entities, err := dockerClient.ContainerList(context.Background(), containerListOptions)
	if err != nil {
		return nil, err
	}
	var selectedContainers []container
	for _, e := range entities {
		c := container{IsContainer: true, Container: e}
		if filter != "" {
			buf, err := json.Marshal(c)
			if err != nil {
				return nil, err
			}

			ok, err := jq.MatchesFilter(string(buf), filter)
			if err != nil {
				return nil, err
			}

			if ok {
				selectedContainers = append(selectedContainers, c)
			}
		} else {
			selectedContainers = append(selectedContainers, c)
		}
	}
	return selectedContainers, nil
}

func deleteContainers(dockerClient *client.Client, containers []container, removeOptions types.ContainerRemoveOptions, killContainerSignal string, stopContainers bool) {
	for _, container := range containers {
		if killContainerSignal != "" || stopContainers {
			details, err := dockerClient.ContainerInspect(context.Background(), container.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to inspect container %s: %s\n", container.ID, err.Error())
				continue
			}
			if details.State.Running {
				if killContainerSignal != "" {
					if err := dockerClient.ContainerKill(context.Background(), container.ID, killContainerSignal); err != nil {
						fmt.Fprintf(os.Stderr, "unable to kill container %s: %s\n", container.ID, err.Error())
						continue
					}
				}
				if stopContainers {
					if err := dockerClient.ContainerStop(context.Background(), container.ID, nil); err != nil {
						fmt.Fprintf(os.Stderr, "unable to stop container %s: %s\n", container.ID, err.Error())
						continue
					}
				}
			}
		}
		if err := dockerClient.ContainerRemove(context.Background(), container.ID, removeOptions); err != nil {
			fmt.Fprintf(os.Stderr, "unable to delete container %s: %s\n", container.ID, err.Error())
		}
	}
}

func selectImages(dockerClient *client.Client, filter string) ([]image, error) {
	entities, err := dockerClient.ImageList(context.Background(), imageListOptions)
	if err != nil {
		return nil, err
	}
	var selectedImages []image
	for _, e := range entities {
		i := image{IsImage: true, ImageSummary: e}
		if filter != "" {
			buf, err := json.Marshal(i)
			if err != nil {
				return nil, err
			}

			ok, err := jq.MatchesFilter(string(buf), filter)
			if err != nil {
				return nil, err
			}

			if ok {
				selectedImages = append(selectedImages, i)
			}
		} else {
			selectedImages = append(selectedImages, i)
		}
	}
	return selectedImages, nil
}

func deleteImages(dockerClient *client.Client, images []image, removeOptions types.ImageRemoveOptions) {
	for _, image := range images {
		if _, err := dockerClient.ImageRemove(context.Background(), image.ID, removeOptions); err != nil {
			fmt.Fprintf(os.Stderr, "unable to delete image %s: %s\n", image.ID, err.Error())
		}
	}
}

func selectNetworks(dockerClient *client.Client, filter string) ([]network, error) {
	entities, err := dockerClient.NetworkList(context.Background(), networkListOptions)
	if err != nil {
		return nil, err
	}
	var selectedNetworks []network
	for _, e := range entities {
		n := network{IsNetwork: true, NetworkResource: e}
		if filter != "" {
			buf, err := json.Marshal(n)
			if err != nil {
				return nil, err
			}

			ok, err := jq.MatchesFilter(string(buf), filter)
			if err != nil {
				return nil, err
			}

			if ok {
				selectedNetworks = append(selectedNetworks, n)
			}
		} else {
			selectedNetworks = append(selectedNetworks, n)
		}
	}
	return selectedNetworks, nil
}

func deleteNetworks(dockerClient *client.Client, networks []network) {
	for _, network := range networks {
		if err := dockerClient.NetworkRemove(context.Background(), network.ID); err != nil {
			fmt.Fprintf(os.Stderr, "unable to delete network %s: %s\n", network.ID, err.Error())
		}
	}
}
