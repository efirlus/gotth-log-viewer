package fsn // fs notify 패키지
// 필요한 모든 핸들링 테크닉, 시간 관리를 작성

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	cmdr "booknotes/modules/commander"
	"booknotes/modules/cv"
	"booknotes/modules/dbpr"
	"booknotes/modules/lg"
	"booknotes/modules/mdcw"
)

type FileState struct { // 파일의 순환 오류를 막기 위한 state struct
	LastModified time.Time   // 마지막 수정 시간
	IsProcessing bool        // 현재 수정중 여부
	IgnoreUntil  time.Time   // 앞으로 차단할 때까지의 '기간'
	Timer        *time.Timer // db 30분 타이머
}

var (
	watcher    *fsnotify.Watcher
	fileStates = make(map[string]*FileState) // [파일명]파일상태, 포인터 사용으로 메모리 오류 방지
	mu         sync.Mutex                    // Protects fileStates, concurrent 확보
)

// 0. watcher 이니시
// 감시할 경로는 cv에 적은 뒤 호출할 것
func InitWatcher() error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("cannot make new watcher: %v", err)
	}

	go WatchFiles()

	err = watcher.Add(cv.MDdir)
	if err != nil {
		return fmt.Errorf("cannot add directory: %v", err)
	}

	err = watcher.Add(cv.DBfile)
	if err != nil {
		return fmt.Errorf("cannot add db file: %v", err)
	}

	return nil
}

//  1. 감시 함수 본체
//     메인이랑 비슷하게 이것도 영구히 수정할 일 없음
func WatchFiles() {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			handleFileEvent(event)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			lg.Err("cannot watch because:", err)
		}
	}
}

// z. fsnotify closer
func CloseWatcher() {
	watcher.Close()
}

// 1-1. routine에 하나밖에 없는 핸들러
func handleFileEvent(event fsnotify.Event) {
	// 파일핸들러 안에서 파일 스테이트를 체크 후 알맞은 케이스로 분류

	// 일단 이벤트.이름, 즉 파일명에서 확장자 추출
	filename := event.Name
	ext := filepath.Ext(filename)

	// ignore events are on 'directory'
	fi, err := os.Stat(filename)
	if err == nil && fi.IsDir() {
		// Ignore directory events
		return
	} else if err != nil {
		lg.Err("cannot recognize event", err)
	}

	mu.Lock()                             // 동시성 선언. 이해 못했음 아직, 이따 자기전에 좀 보자
	state, exists := fileStates[filename] // filestates 맵에 파일명 키가 있나 확인
	if !exists {                          // 없으면 값으로 들어갈 스테이트가 빈 키 생성
		state = &FileState{}
		fileStates[filename] = state
	}
	mu.Unlock()

	// 1. Check if we should ignore this event
	if time.Now().Before(state.IgnoreUntil) {
		return
	}
	// 2. Skip if already processing
	if state.IsProcessing {
		return
	}
	// 3. 여기까지가 뭐냐면, 일단 ignoreuntil 상태거나 isprocessing 상태면
	// 아예 아래 함수를 부르지 말고 바로 탈출하란 뜻이고
	// 앞으로 처리해야 될 상황이 맞으면
	state.IsProcessing = true // isprocessing을 트루로 바꾼 뒤
	defer func() {            // 아래의 케이스 함수가 다 끝난 뒤에
		state.IsProcessing = false // isprocessing을 다시 false로 변경하란 뜻
	}()

	// 그리고 확장자 별로 함수 호출
	switch ext {
	case ".md":
		handleMdFile(filename, event)
	case ".db":
		handleDbFile(filename, event)
	}

	filename = ""
}

// 1-1-a. md파일 핸들러
func handleMdFile(filename string, event fsnotify.Event) {
	// 삭제 경우의 수는 상정 안하기로 결정, 너무 어렵기도 어렵거니와
	// 에러 위험성이 너무 상존함, 나의 구멍 뚫린 실력으론 조금;;
	// 1. 파일이 생성/수정될 때
	if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
		lg.Info(fmt.Sprintf("Processing .md file: %s", filename))
		// 파일의 close-write 임시 구현
		mu.Lock()
		state := fileStates[filename]

		if state.Timer != nil {
			state.Timer.Stop()
		}

		state.Timer = time.AfterFunc(20*time.Second, func() { // 10초
			// 이 자리 알고리즘
			ignoreFile(cv.DBfile, 30*time.Second)

			mdcw.MDCreateWrite(filename)
		})
	}
}

// 1-1-b. db 파일 감시기기
func handleDbFile(filename string, event fsnotify.Event) {
	// md쪽 이벤트 핸들러가 작동한 지 30분이 지난 뒤에만 db쪽 핸들러가 작동되게 하는 것
	// md이벤트들이 지속적으로 db파일을 수정할 것이기 때문에 순환오류를 막으려면 반드시 타이머가 필요함
	// 경우의 수 하나
	if event.Op&fsnotify.Write != 0 {
		lg.Info("DB file updated")
		// Your logic for handling .db file update here

		mu.Lock()
		state := fileStates[filename]

		//reset timer if timer is exist
		if state.Timer != nil {
			state.Timer.Stop()
		}

		//set a new timer
		state.Timer = time.AfterFunc(30*time.Minute, func() {
			// AfterFunc가 헷갈리는데.. 하여튼 이게 대충 무슨 내용이냐면
			// time 30분동안 대기를 할거야, 근데 그 대기시간은 이 함수가 반복되는 동안 이 윗 구절 (7줄 위)때문에 계속 리셋될거야
			// 그러다가 시간이 리셋 안된지 30분이 흐르면 func()를 실행해, 그 func 내용이 여기 들어가야되는데
			// 다 적지 않고 밖으로 뺀거지
			processDbFile()
			// 그 다음에, 이 func가 시행되는 시점을 timer가 새롭게 세팅하는거야
		})
		//타이머가 세팅되고 나면 바로 mu 언락
		mu.Unlock()
	}
}

// 1-1-b-a. db파일 기반 처리기
func processDbFile() {
	lg.Info("Processing DB file after 30 minutes of inactivity")

	// list db -> directory walk -> compare
	//     -> delete md -> return notYet까지
	idMap := dbpr.DBhandleFirst()

	// create md 는 중간에 ignore를 껴야 돼서
	// 함수로는 못 묶음
	for id, title := range idMap {

		rcmd := cmdr.ShowMD(id)
		bres, err := cmdr.CommandExec(rcmd)
		if err != nil {
			lg.Err("cannot execute show_metadata command", err)
		}

		parsedResult, err := dbpr.ParseResult(bres)
		if err != nil {
			lg.Err("cannot parse metadata from command result", err)
		}

		fullfilepath := filepath.Join(cv.MDdir, title+".md")

		ignoreFile(fullfilepath, 30*time.Second)

		_, err = dbpr.GenerateMarkdown(parsedResult)
		if err != nil {
			lg.Err("cannot generate markdown of book", err)
		}

		rcmd = nil
		bres = ""
		parsedResult = dbpr.BookMetadata{}
		fullfilepath = ""
	}

	idMap = nil
}

// 이벤트 무시 시간 설정 함수
func ignoreFile(filename string, duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()

	state, exists := fileStates[filename]
	if !exists {
		state = &FileState{}
		fileStates[filename] = state
	}

	state.IgnoreUntil = time.Now().Add(duration)
}
