## 자동 북노트 메이커

---
Created: 2024-09-04 13:11
tags:
  - 🌶️
분류:
  - "[[데이터뷰]]"
  - "[[서버 관리]]"
  - "[[소설]]"
  - "[[워크플로우 자동화]]"
---

### 체크리스트
- [x] [[Calibre Server]]에서 자체지원하는 calibredb - cli 명령어를 통해 db내의 모든 도서 정보를 긁어옴
- [x] [[책 - 관심|책 템플릿]]에 걸맞는 형식으로 긁어온 정보를 [[마크다운]] 문서화
- [x] 하는 동시에 각 노트의 필수 정보 (id, title, 문서내용기반md5) 를 binary 파일로 저장
- [x] 생성된 문서는 [[OneDrive for Linux]]를 통해 업로드
- [x] 문서의 modified 여부를 감시하기 위해 계속 md5 체크
- [x] 수정된 내용을 서버에 반영하기 위해 마크다운을 해석
- [x] 해석된 내용을 cli로 입력
- [x] [[Calibre]]가 자체지원하는 `calibre://` 형식 링크와 `[[]]` 상호변환, 내용에 따라 링크 자동 생성


### 업데이트 내역

###### [[2024-10-08]] 11:27 추가
- rune은 메모리 효율이 떨어지고 느리다고 함, fsnotify를 돌리고 있다면 느린게 치명적임. 그래서 안 돌리기로 함
###### [[2024-09-29]] 13:37 추가
- 씨발 ㅋㅋㅋㅋㅋㅋㅋㅋㅋㅋㅋ main.go 통째로 날림 ㅋㅋㅋㅋㅋㅋㅋㅋㅋㅋㅋㅋㅋㅋㅋ
- 아 너무 빡친다 ㅅㅂ;; 이참에 sqlite3 위주로 짜볼까 싶기도 하고...아냐 그냥 현재 폼을 유지하자 너무 나갔다 그건
- go 모듈 중 하나인 fsnotify로 메카니즘 바꿔볼 예정, create, modify, delete 이벤트에 대해 해당 파일명-id 바이너리를 관리하도록 하면 될 듯
###### [[2024-09-28]] 20:37 추가
- 아 ㅅㅂ ㅋㅋㅋㅋㅋㅋㅋㅋㅋ readbinary 함수가 md5만 두번 저장하네 미친 ㅋㅋㅋㅋㅋㅋㅋㅋㅋ 이거 다 좆된거 아닌가 싶다;;일단 테스트 해봐야지
- 
###### [[2024-09-28]] 20:27 추가
- 주소가 이상하게 생성되네... 알고리즘 까봐야 함
- 분명히 다 db에 저장된 책 '제목'인데 제대로 못 읽고 전부다 author 링크 처리해버림 ㅅㅂ;;;
```html
<div>
<p>영국을 기반으로 한 소설들 중에 그나마 제일 재밌게 읽은 건 <a href="calibre://show-note/Book/authors/hex_ecb29ceba788ed9988eca68820eb9fb0eb8d98ec9599ebb3b5">천마홈즈 런던앙복</a>이다</p>
<p>뭐 <a href="calibre://show-note/Book/authors/hex_eb8c80ec9881eca09ceab5adec9d9820ed95a8ec9ea5ec9db420eb9098ec9788eb8ba4">대영제국의 함장이 되었다</a> 라던지, <a href="calibre://show-note/Book/authors/hex_eb8c80ec9881ecb29ced95982c20eca1b0ec84a0eba78cec84b8">대영천하, 조선만세</a> 도 있긴한데, 앞에 건 내 기억이 왜곡된 건지 취향이 바뀐건지 몰라도 하여튼 그렇고, 뒤에건 연재가 ㅅㅂ</p>
<p>하여튼, 영국이 메인인 매체의 관점은 한결같은 면이 있다. 식민주의, 혐성, 패배주의</p>
<p>이번 소설엔 그런 뉘앙스가 상당히 적다는 게 매력이다</p>
<p>단점은 주인공이 착각물 치곤 너무 일잘러이고, 근성/흑막 치곤 너무 등신이다. <a href="calibre://show-note/Book/authors/hex_ebaa85eab5b0ec9db420eb9098ec96b4ebb3b4ec84b82031ebb680202d20ebacb4eca285">명군이 되어보세 1부 - 무종</a> 이후 주인공이 열심히 하는 병신인 장르가 늘긴 했지...</p>
<p>주변 인물이 평면적인 것도 단점, <a href="calibre://show-note/Book/authors/hex_ebacb4ec84a020ec97b0eab2b020ec98a4eb8298ed9980eba19c20eb94b0eba8b9eab8b0">무선 연결 오나홀로 따먹기</a> 같이 보지에 떡만 치는 소설도 이거보단 입체적</p>
<p>희귀한 컨셉이라 250화 기준 별 4개를 줬지만, 과연 연재 완료 이후 내 평가가 바뀔까?</p></div>
```
###### [[2024-09-25]] 16:24 추가
- 링크 생성 기능 추가함, 태그화는 약간 애매하고, 마크다운화는 완벽하지 뭐
###### [[2024-09-24]] 16:43 추가
- calibre://book-details/_hex_-426f6f6b/682  이런 형식의 링크로 코멘트 달기
- `<a href="calibre://book-details/_hex_-426f6f6b/682">이런 형식</a>` 으로 변환하는 메커니즘을 넣으면 됨. 반대도 마찬가지. a href 태그를 삭제하고 `[[]]` 로 변환해 주는 기능을 넣는거지. 뭐... 라이브러리 코드는 하드코딩한다 치고. 
- [작가 노트 링크](calibre://show-note/Book/authors/hex_6275726e38)가 되긴 하는데, 작가노트는 gui 클라이언트에서만 생성 가능하고, 아직 cli에 포함이 안됐다는 치명적인 문제가 있네 ㅋㅋㅋㅋㅋㅋㅋㅋ

###### [[2024-09-05]] 11:08 추가
- [[튜토리얼 라이프]] : - 세팅 과정에서 충돌이 일어나서 663 아이디 삭제 후 1007로 재등록함... 근본적 원인은 205랑 663이 같이 있어서 생긴 문제였음, 앞으로는 중복 제목은 주의할 것.
- 이를 개선하기 위해 제목에 - id를 붙이는 걸 진지하게 고민해봐야 하는데, 근데 그러면 노트제목 - # 둘을 동기화를 못한다는 치명적인 문제가 생김
### 개요
- crontab으로 `gobooknotes 패스코드`를 치면 매시 15분, 45분에 저절로 실행됨
- 생성된 책 노트들은 dataview로 자동 목록화
- 문제:: 한글과 이모지, 유니코드의 능수능란하고 오류없는 지원을 위해 rune 도입을 시도할 것
- 이걸 하면 아마도 linebreak 오류는 사라질 듯
### 설명서
#### 책 상태 - $status
- "안읽음" = "📘" `:blue_book`
- "읽는중" = "📖" `:book`
- "읽음" = "📗" `:green_book`
- "읽다맘" = "📕" `:closed_book`
- "대기중" = "🔖" `:bookmark`

### 알고리즘 재생성

1. ob에서 문서를 수정
	1. md5 확인 방식으로 수정된 문서 체크
	2. 
2. ob에서 문서 삭제
	1. 이것도 체크할 일 없음
3. ca에서 문서 수정
	1. 어차피 체크할 일 없을 듯
4. ca에서 문서 삭제
	1. ca에서 빠진 id 확인
	2. 해당 id에 해당하는 title 체크
	3. title file delete
	4. binary 에서 제거
5. ca에서 문서 추가
	1. ca에서 추가된 문서 id 겟
	2. 해당 id로 show metadata
	3. 얻어진 md -> 문서화
	4. 바이너리 추가

### 재작업기 - [[2024-09-29]]

```go
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if strings.HasSuffix(event.Name, ".md") {
					if event.Op&fsnotify.Create == fsnotify.Create ||
						event.Op&fsnotify.Write == fsnotify.Write {
						handleCreateWrite(event.Name)
						lastEventTime = time.Now()
					} else if event.Op&fsnotify.Remove == fsnotify.Remove {
						handleDelete(event.Name)
						lastEventTime = time.Now()
					}
				}

				// Handle single ".db" file (write only)
				if strings.HasSuffix(event.Name, ".db") {
					if event.Op&fsnotify.Write == fsnotify.Write {
						handleDbEvent(event)
						lastEventTime = time.Now()
					}
				}

				resetTimer(lastEventTime, timeout)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	<-done
```
