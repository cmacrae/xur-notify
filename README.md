[![Go Report Card](https://goreportcard.com/badge/github.com/cmacrae/xur-notify)](https://goreportcard.com/report/github.com/cmacrae/xur-notify) [![Build Status](https://travis-ci.org/cmacrae/xur-notify.svg?branch=master)](https://travis-ci.org/cmacrae/xur-notify)
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
In this current implementation, `xur-notify` requires the following environment variables:  
- `BNET_API_KEY` - the following value should be used: `48cfb02afcac408aaa49586aba482cf9`
- `PUSHOVER_TOKEN` - the following value should be used: `ab7wkogn9xqixemty1iei8jdb7uebw`
- `PUSHOVER_USER_KEY` - use your user token, from [pushover.net](https://pushover.net)

### Docker
A [Docker image (`cmacrae/xur-notify`)](https://hub.docker.com/r/cmacrae/xur-notify/) is available, and should be run like so:
``` bash
docker run --name xur-notify -d -e PUSHOVER_TOKEN=$PUSHOVER_TOKEN -e PUSHOVER_USER_KEY=$PUSHOVER_USER_KEY -e BNET_API_KEY=$BNET_API_KEY cmacrae/xur-notify:1.0.0-alpha
```
_Note: the relevant environment variables would need to be set in order for this to work_

This Docker image simply uses `cron` to execute `xur-notify` at 10:00 AM UTC (Xûr's arrival time).  
If you'd like to alter this to suit your schedule, you can build your own Docker image changing the following line to your own time in `cron` format:  
``` cron
RUN echo -e "0\t10\t*\t*\t5\t/bin/xur-notify &> /dev/stdout" > /etc/crontabs/root
```

### Binary
Or, simply head over to [the releases page](https://github.com/cmacrae/xur-notify/releases) and grab the latest binary for your platform, set your environment variables, and run it!  
You'll probably want to have this execute on a schedule, so you get seemless notifications when Xûr arrives.

## Future implementations (less hands-on)
This is the first program I've written.  
Right now, it does a great job at doing what I set out to do, but leaves some desire in its means of deployment for end users.  
I plan to make this more consumer friendly as the project evolves, hoping to make it easy for fellow Guardians to get the info they want, easily.  

So, hang tight, if you're not so comfortable with all the above!
