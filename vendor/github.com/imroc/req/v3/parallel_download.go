package req

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ParallelDownload struct {
	url          string
	client       *Client
	concurrency  int
	output       io.Writer
	filename     string
	segmentSize  int64
	perm         os.FileMode
	tempRootDir  string
	tempDir      string
	taskCh       chan *downloadTask
	doneCh       chan struct{}
	wgDoneCh     chan struct{}
	errCh        chan error
	wg           sync.WaitGroup
	taskMap      map[int]*downloadTask
	taskNotifyCh chan *downloadTask
	mu           sync.Mutex
	lastIndex    int
}

func (pd *ParallelDownload) completeTask(task *downloadTask) {
	pd.mu.Lock()
	pd.taskMap[task.index] = task
	pd.mu.Unlock()
	go func() {
		select {
		case pd.taskNotifyCh <- task:
		case <-pd.doneCh:
		}
	}()
}

func (pd *ParallelDownload) popTask(index int) *downloadTask {
	pd.mu.Lock()
	if task, ok := pd.taskMap[index]; ok {
		delete(pd.taskMap, index)
		pd.mu.Unlock()
		return task
	}
	pd.mu.Unlock()
	for {
		task := <-pd.taskNotifyCh
		if task.index == index {
			pd.mu.Lock()
			delete(pd.taskMap, index)
			pd.mu.Unlock()
			return task
		}
	}
}

func md5Sum(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func (pd *ParallelDownload) ensure() error {
	if pd.concurrency <= 0 {
		pd.concurrency = 5
	}
	if pd.segmentSize <= 0 {
		pd.segmentSize = 1073741824 // 10MB
	}
	if pd.perm == 0 {
		pd.perm = 0777
	}
	if pd.tempRootDir == "" {
		pd.tempRootDir = os.TempDir()
	}
	pd.tempDir = filepath.Join(pd.tempRootDir, md5Sum(pd.url))
	if pd.client.DebugLog {
		pd.client.log.Debugf("use temporary directory %s", pd.tempDir)
		pd.client.log.Debugf("download with %d concurrency and %d bytes segment size", pd.concurrency, pd.segmentSize)
	}
	err := os.MkdirAll(pd.tempDir, os.ModePerm)
	if err != nil {
		return err
	}

	pd.taskCh = make(chan *downloadTask)
	pd.doneCh = make(chan struct{})
	pd.wgDoneCh = make(chan struct{})
	pd.errCh = make(chan error)
	pd.taskMap = make(map[int]*downloadTask)
	pd.taskNotifyCh = make(chan *downloadTask)
	return nil
}

func (pd *ParallelDownload) SetSegmentSize(segmentSize int64) *ParallelDownload {
	pd.segmentSize = segmentSize
	return pd
}

func (pd *ParallelDownload) SetTempRootDir(tempRootDir string) *ParallelDownload {
	pd.tempRootDir = tempRootDir
	return pd
}

func (pd *ParallelDownload) SetFileMode(perm os.FileMode) *ParallelDownload {
	pd.perm = perm
	return pd
}

func (pd *ParallelDownload) SetConcurrency(concurrency int) *ParallelDownload {
	pd.concurrency = concurrency
	return pd
}

func (pd *ParallelDownload) SetOutput(output io.Writer) *ParallelDownload {
	if output != nil {
		pd.output = output
	}
	return pd
}

func (pd *ParallelDownload) SetOutputFile(filename string) *ParallelDownload {
	pd.filename = filename
	return pd
}

func getRangeTempFile(rangeStart, rangeEnd int64, workerDir string) string {
	return filepath.Join(workerDir, fmt.Sprintf("temp-%d-%d", rangeStart, rangeEnd))
}

type downloadTask struct {
	index                int
	rangeStart, rangeEnd int64
	tempFilename         string
	tempFile             *os.File
}

func (pd *ParallelDownload) handleTask(t *downloadTask, ctx ...context.Context) {
	pd.wg.Add(1)
	defer pd.wg.Done()
	t.tempFilename = getRangeTempFile(t.rangeStart, t.rangeEnd, pd.tempDir)
	if pd.client.DebugLog {
		pd.client.log.Debugf("downloading segment %d-%d", t.rangeStart, t.rangeEnd)
	}
	file, err := os.OpenFile(t.tempFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		pd.errCh <- err
		return
	}
	err = pd.client.Get(pd.url).
		SetHeader("Range", fmt.Sprintf("bytes=%d-%d", t.rangeStart, t.rangeEnd)).
		SetOutput(file).
		Do(ctx...).Err

	if err != nil {
		pd.errCh <- err
		return
	}
	t.tempFile = file
	pd.completeTask(t)
}

func (pd *ParallelDownload) startWorker(ctx ...context.Context) {
	for {
		select {
		case t := <-pd.taskCh:
			pd.handleTask(t, ctx...)
		case <-pd.doneCh:
			return
		}
	}
}

func (pd *ParallelDownload) mergeFile() {
	defer pd.wg.Done()
	file, err := pd.getOutputFile()
	if err != nil {
		pd.errCh <- err
		return
	}
	for i := 0; ; i++ {
		task := pd.popTask(i)
		tempFile, err := os.Open(task.tempFilename)
		if err != nil {
			pd.errCh <- err
			return
		}
		_, err = io.Copy(file, tempFile)
		tempFile.Close()
		if err != nil {
			pd.errCh <- err
			return
		}
		if i < pd.lastIndex {
			continue
		}
		break
	}
	if pd.client.DebugLog {
		pd.client.log.Debugf("removing temporary directory %s", pd.tempDir)
	}
	err = os.RemoveAll(pd.tempDir)
	if err != nil {
		pd.errCh <- err
	}
}

func (pd *ParallelDownload) Do(ctx ...context.Context) error {
	err := pd.ensure()
	if err != nil {
		return err
	}
	for i := 0; i < pd.concurrency; i++ {
		go pd.startWorker(ctx...)
	}
	resp := pd.client.Head(pd.url).Do(ctx...)
	if resp.Err != nil {
		return resp.Err
	}
	if resp.ContentLength <= 0 {
		return fmt.Errorf("bad content length: %d", resp.ContentLength)
	}
	pd.lastIndex = int(math.Ceil(float64(resp.ContentLength)/float64(pd.segmentSize))) - 1
	pd.wg.Add(1)
	go pd.mergeFile()
	go func() {
		pd.wg.Wait()
		close(pd.wgDoneCh)
	}()
	totalBytes := resp.ContentLength
	start := int64(0)
	for i := 0; ; i++ {
		end := start + (pd.segmentSize - 1)
		if end > (totalBytes - 1) {
			end = totalBytes - 1
		}
		task := &downloadTask{
			index:      i,
			rangeStart: start,
			rangeEnd:   end,
		}
		pd.taskCh <- task
		if end < (totalBytes - 1) {
			start = end + 1
			continue
		}
		break
	}
	select {
	case <-pd.wgDoneCh:
		if pd.client.DebugLog {
			if pd.filename != "" {
				pd.client.log.Debugf("download completed from %s to %s", pd.url, pd.filename)
			} else {
				pd.client.log.Debugf("download completed for %s", pd.url)
			}
		}
		close(pd.doneCh)
	case err := <-pd.errCh:
		return err
	}
	return nil
}

func (pd *ParallelDownload) getOutputFile() (io.Writer, error) {
	outputFile := pd.output
	if outputFile != nil {
		return outputFile, nil
	}
	if pd.filename == "" {
		u, err := urlpkg.Parse(pd.url)
		if err != nil {
			panic(err)
		}
		paths := strings.Split(u.Path, "/")
		for i := len(paths) - 1; i > 0; i-- {
			if paths[i] != "" {
				pd.filename = paths[i]
				break
			}
		}
		if pd.filename == "" {
			pd.filename = "download"
		}
	}
	if pd.client.outputDirectory != "" && !filepath.IsAbs(pd.filename) {
		pd.filename = filepath.Join(pd.client.outputDirectory, pd.filename)
	}
	return os.OpenFile(pd.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, pd.perm)
}
