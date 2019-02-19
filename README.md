# docker-purge
Delete docker images with jq.

# Usage
```
usage: docker-purge [<flags>] [<filter>]

Flags:
      --help                    Show context-sensitive help (also try
                                --help-long and --help-man).
      --list-all                list docker containers, images, networks
      --list-containers         list docker containers
      --list-images             list docker images
      --list-networks           list docker networks
  -d, --dry                     dry run, do not purge anything
      --containers              limit purge to docker containers
      --images                  limit purge to docker images
      --networks                limit purge to docker networks
      --force                   sets container.remove.force and
                                image.remove.force to true
      --all                     remove everything related to an entity
      --container.remove.force  force removal of container
      --container.remove.links  remove links during removal
      --container.remove.volumes  
                                remove volumes during removal
      --container.stop          stop running docker container
      --container.kill=""       kill running docker container with the specified
                                signal
      --image.remove.force      force removal of image
      --image.remove.prunechildren  
                                prune children on removal

Args:
  [<filter>]  jq filter to apply
```
## Examples

Delete all containers that have firefox in their name
```bash
docker-purge '.IsContainer == true and (.Image | contains("firefox"))'
```

## Notice
Building is more less broken...  
Try to run `make build` and see if it generates a dist/ for you
