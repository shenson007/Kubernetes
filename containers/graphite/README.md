# Docker Image for Graphite

## Get Graphite running instantly to use with sitespeed.io

This is a fork from [Hopsoft](https://github.com/hopsoft/docker-graphite-statsd) Massive love to Hopsoft for setting up the original image.

In this image, Graphite is always setup with Basic Auth (feed your .htpasswd file when starting) and the Graphite data dir is set to */opt/graphite/storage/whisper*.

## Quick Start

```sh
sudo docker run -d \
  --name graphite \
  -p 8080:80 \
  -p 2003:2003 \
  sitespeedio/graphite
```

This starts a Docker container named: **graphite** with Basic Auth **guest/guest**. Please change the login that by feeding your own .htpasswd file when starting the container ([more info about how to create your .htpasswd file](http://httpd.apache.org/docs/2.2/programs/htpasswd.html)):

```sh
sudo docker run -d \
  --name graphite \
  -p 8080:80 \
  -p 2003:2003 \
  -v /local/path/to/.htpasswd:/etc/nginx/.htpasswd \
  sitespeedio/graphite
```

And the final config that you should do is map the Graphite data dir outside of your container:

```sh
sudo docker run -d \
  --name graphite \
  -p 8080:80 \
  -p 2003:2003 \
  -v /local/path/to/.htpasswd:/etc/nginx/.htpasswd \
  -v /path/to/data/graphite/storage/whisper:/opt/graphite/storage/whisper \
  sitespeedio/graphite
```

TODO also map log dirs

## Data retention
You can change how often data will be stored in the  [storage-schemas.conf](https://github.com/sitespeedio/docker-graphite-statsd/blob/master/conf/graphite/storage-schemas.conf) and how metrics will be aggregated over time in [storage-aggregation.conf](https://github.com/sitespeedio/docker-graphite-statsd/blob/master/conf/graphite/storage-aggregation.conf).

The default one looks like this:

```
retentions = 5m:1d,15m:21d,30m:60d
```

It will store data for 2 months, change that if you need to store data longer. Etsy has good [documentation](https://github.com/etsy/statsd/blob/master/docs/graphite.md) on how to setup your Graphite metrics.

To change it, you can feed the image with a new *storage-schemas.conf*. The one you want to replace is located  
*/opt/graphite/conf/storage-schemas.conf*

Start your image like this:

```sh
sudo docker run -d \
  --name graphite \
  -p 8080:80 \
  -p 2003:2003 \
  -v /local/path/to/.htpasswd:/etc/nginx/.htpasswd \
  -v /path/to/data/graphite/storage/whisper:/opt/graphite/storage/whisper \
  -v /path/to/storage-schemas.conf:/opt/graphite/conf/storage-schemas.conf \
  sitespeedio/graphite
```


### Base Image

Built using [Phusion's base image](https://github.com/phusion/baseimage-docker).

* All Graphite related processes are run as daemons & monitored with [runit](http://smarden.org/runit/).
* Includes additional services such as logrotate.
