[![Go Report Card](https://goreportcard.com/badge/github.com/cmacrae/xur-notify)](https://goreportcard.com/report/github.com/cmacrae/xur-notify) [![Build Status](https://travis-ci.org/cmacrae/xur-notify.svg?branch=master)](https://travis-ci.org/cmacrae/xur-notify) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](LICENSE)
# xur-notify
Push notifications for Xûr's inventory

## About
Xûr is the mysterious traveling salesman in [Bungie](https://bungie.net)'s game [Destiny](https://www.destinythegame.com/uk/en/home). Each weekend he visits, he brings with him exotic items which he puts up for sale for Guardians.  

`xur-notify` is a little utility, intended to be run on schedule, to instantly notify Guardian's of Xûr's arrival and the particular gifts he bears.

Currently, the following notification methods are supported:  
- Push notifications, with the excellent [Pushover](https://pushover.net)  
_Note: plans for other notification methods are emerging_

## What's in the notification
Right now, the notification simply sends a list of the exotic armour/weapons on sale.  
In future versions, more detailed information will be provided, such as stat-roll and quality.

## Preview
Lockscreen                 |  Pushover
:-------------------------:|:-------------------------:
![](http://i.imgur.com/j1YJwSN.png)  |  ![](http://i.imgur.com/l6dbZq4.png)

## Usage
[**Subscribe via Pushover**](https://pushover.net/subscribe/Xr-mwxq4o1v35qs8er) to receive notifications from `xur-notify` at 10:00AM UTC every Friday!  

_Warning: This is an experimental implementation - due to potential influx in users, please bear with me in this initial release. I'm personally hosting the infrastructure triggering the notifications, so there may be some unforeseen problems in early usage_

### Running your own instance of xur-notify
I have Open Source software and the brilliant community behind it to thank for so much in my life, I'd like to stay true to my roots.  
So, if you'd rather just run your own `xur-notify`, feel free!  

In this current implementation, `xur-notify` requires the following environment variables:  
- `BNET_API_KEY` - can be obtained from [bungie.net](https://www.bungie.net/en/Application)
- `PUSHOVER_TOKEN` - use your own application token, from [pushover.net](https://pushover.net)
- `PUSHOVER_RECIPIENT_KEY` - use your user/group token, from [pushover.net](https://pushover.net)
- `TIMEZONE` - specify your timezone, defaults to `Europe/London` if unset (use tz database zone format)

#### Docker
A [Docker image (`cmacrae/xur-notify`)](https://hub.docker.com/r/cmacrae/xur-notify/) is available, and should be run like so:
``` bash
docker run --name xur-notify -d -e PUSHOVER_TOKEN=$PUSHOVER_TOKEN -e PUSHOVER_RECIPIENT_KEY=$PUSHOVER_RECIPIENT_KEY -e BNET_API_KEY=$BNET_API_KEY cmacrae/xur-notify:1.0.0-alpha
```
_Note: the relevant environment variables would need to be set in order for this to work_

This Docker image simply uses `cron` to execute `xur-notify` at 10:00 AM UTC (Xûr's arrival time).  
If you'd like to alter this to suit your schedule, you can build your own Docker image changing the following line to your own time in `cron` format:  
``` cron
RUN echo -e "0\t10\t*\t*\t5\t/bin/xur-notify &> /dev/stdout" > /etc/crontabs/root
```

#### Binary
Or, simply head over to [the releases page](https://github.com/cmacrae/xur-notify/releases) and grab the latest binary for your platform, set your environment variables, and run it!  
You'll probably want to have this execute on a schedule, so you get seemless notifications when Xûr arrives.

## Licensing
See the [LICENSE](LICENSE) file

## Credit
- @gregdel: [for making Pushover integration so slick](https://github.com/gregdel/pushover)
- @jmoiron: [for easing handling of json structures](https://github.com/jmoiron/jsonq)
- The awesome people at [superblock](https://superblock.net/contact/) for [Pushover](https://pushover.net)!
- [Allyn H](http://allynh.com/blog/) for his series of blog posts that inspired me to write `xur-notify`
