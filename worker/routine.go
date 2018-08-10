package worker

import (
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/api"
)

func (c *workerContext) workRoutine() {
	for {
		// Check if routine should exit
		select {
			case <-c.ctxt.Done(): break
			default:
		}

		// TODO Move video back to wait queue if processing failed

		videoId := <-c.jobs
		req := apis.Main.GrabVideo(videoId)
		res, err := net.Client.Do(req)
		if err != nil {
			log.Errorf("Failed to download video \"%s\": %s", videoId, err.Error())
			c.errors <- err
		}

		var v data.Video
		v.ID = videoId
		var result interface{}

		next, err := apis.Main.ParseVideo(&v, res)
		if err == api.VideoUnavailable {
			log.Debugf("Video is unavailable: %s", videoId)
			result = data.CrawlError{
				VideoId: videoId,
				Err: api.VideoUnavailable,
				VisitedTime: time.Now(),
			}
		} else if err != nil {
			log.Errorf("Parsing video \"%s\" failed: %s", videoId, err.Error())
			c.errors <- err
		} else {
			result = data.Crawl{
				Video: &v,
				VisitedTime: time.Now(),
			}
		}

		c.results <- result

		if len(next) > 0 {
			for _, id := range next {
				c.newIDs <- id
			}
		}
	}
}