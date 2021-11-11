package process

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

const (
	NUM_CONCURRENT = 50 // 同時に処理できるファイル数
)

func Walk(src, dst string, skipper Skipper, processor Processor) error {
	errs := make(chan error, NUM_CONCURRENT)
	lock := make(chan struct{}, NUM_CONCURRENT)
	passedAll := make(chan struct{})
	stopWalking, totalErr := handleProcesses(errs, lock, passedAll)
	var wg sync.WaitGroup

	// walk を抜けるのは, ↓の2通り
	// 1. walk 中にエラーが発生しなかった -> totalErr に nil が送信されている
	// 2. walk 中にエラーが発生した -> totalErr にエラーが送信されている
	err := filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		rpath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if skipper.Skip(rpath) {
			return filepath.SkipDir
		}

		newpath := filepath.Join(dst, rpath)
		if info.IsDir() {
			if _, err := os.Stat(newpath); !os.IsNotExist(err) {
				return nil
			}
			return os.Mkdir(newpath, 0o777)
		}

		select {
		case <-stopWalking:
			return filepath.SkipDir
		case lock <- struct{}{}:
			wg.Add(1)
			go func() {
				errs <- processor.Process(path, newpath)
				wg.Done()
			}()
		}
		return nil
	})

	// walk の終了を handleProcesses に伝える
	wg.Wait()
	close(passedAll)

	if err != nil {
		return err
	}

	return <-totalErr
}

// 正常終了 -> senderr に nil を返す
// 異常終了 -> senderr に エラーを格納 & stop チャネルを閉じる
func handleProcesses(geterr <-chan error, lock chan struct{}, passedAll <-chan struct{}) (stopWalking chan struct{}, senderr chan error) {
	senderr = make(chan error)
	stopWalking = make(chan struct{})
	go func() {
		for {
			select {
			case <-lock:
			case <-passedAll: // すべてのディレクトリの walk が終了したら, return
				senderr <- nil
				return
			}

			err := <-geterr
			if err != nil {
				close(stopWalking) // エラーをチャネルに流すより先に close しておかないと, ブロックしてしまう
				senderr <- err
				return
			}
		}
	}()
	return stopWalking, senderr
}
