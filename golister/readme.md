

### 체크리스트
- [x] 미디어가 저장된 폴더에서 미디어 파일의 목록을 쭉 읽어들임
	- [x] mmd
	- [x] pmv
	- [x] fancam
	- [x] av
	- [ ] h-ani
- [x] [[PowerShell Script]]를 통해 윈도우와 동기화중인 dpl 파일에서 마지막 재생 위치를 확인
- [x] [[모바일 디바이스 액세스 체커]]로는 컴퓨터가 아니라 탭/폰으로 액세스한 마지막 파일을 확인
- [x] 확인된 마지막 파일을 기준으로, '이미 본' 것은 유지하되 디렉토리에서 삭제된 파일들의 항목은 리스트에서도 제거
- [x] '아직 안 본' 것은 무작위 셔플
	- [x] 단 mmd는 디렉토리만 셔플 - 파일은 이름명 순으로
	- [x] av는 "같은 시리즈의 나눔파일"은 연속되게, "같은 레이블"은 불연속되게
	- [ ] 애니는 한 시리즈 연속, 그 외 랜덤
	- [x] 나머지는 완전랜덤 정렬
- [x] 셔플로 만들어진 새로운 재생목록을 각각 pc용 재생목록 (이미 본것 + 아직 안본것) 과
- [x] 모바일용 재생목록 (아직 안본것) 으로 나누어 생성
- [x] 알고리즘을 맵타입으로 변경
### 업데이트 내역
- 2024-08-22 : v1.0
- 2024-10-30 : v2.0 - 3개랜덤, 각 랜덤 알고리즘의 정합성, 로그 개선

### 개요
- 3가지 서로 다른 랜덤 함수를 스위치케이스로 적용해 m3u를 만듬
- m3u 역시 pc용과 모바일용을 구분-생성해줌
- 각 재생목록은 가장 최신으로 본 부분까지를 자동으로 동기화해 다음 부분을 새로 믹스해줌
- 



### 설명서

#### 명령어 바로가기
```cmd
go build -o golister main.go
```
> -o 인수는 프로그램명을 만들어줌

```cmd
sudo ln -s /home/efirlus/gommdlisters/golister /usr/local/bin/golister
```
> 홈경로를 다 적어야 하는 이유는 그냥 gomm/golis 하면 심링크->심링크로 처리되기 때문에 execution 명령어 작동이 안됨. 직접 만든 프로그램은 반드시 fullpath로 생성시킬것

#### 명령어 인수 받게 만들기
```GoLang
// 만약 명령어가 goprogram exe1 exe2 라면
os.Args[1] = exe1
os.Args[2] = exe2

//따라서 입력받고 싶은 variable에다가 var1 := os.Args[n] 하면 프로그램적 처리가 가능
```

#### 상시 가동을 위해 systemctl service로 만들기
``` 
// 파일명은 : /etc/systemd/system/golister-watcher@.service
[Unit]
Description=Service to run golister based on file modification (%i)
StartLimitIntervalSec=10
StartLimitBurst=5

[Service]
Type=oneshot
ExecStart=/bin/bash -c '/etc/systemd/system/golister-handler.sh %i'

[Install]
WantedBy=multi-user.target
```
>키포인트는 파일명에 @, 이하의 watcher.path들이 여러개라면 @을 통해 구분해야 하는데, 이 때 각 path들이 자동으로 이 서비스를 실행시킬 수 있도록 하려면 @가 필수. 이렇게 만듦으로서 실질적으로 watcher 서비스는 비가동상태를 유지하고 있다가 path가 요청할 때만 가동하는 방식

```
// 파일명은 : /etc/systemd/system/golister-watcher@NAS3-samba-watch-fancam.path

[Unit] 
Description=Watch /NAS3/samba/watch/fancam.dpl for changes 

[Path] 
PathChanged=/NAS3/samba/watch/fancam.dpl 

[Install] 
WantedBy=multi-user.target
```
> 예시 하나만 이렇게 제시하는데 다른 2 파일도 마찬가지 구조임, 굉장히 간단하고, path changed만 있음. 그러면 대체 프로그램 실행에 대한 코드는 어딨냐, 없음, @ 로 연결되어 있기 때문에 각 path 하나하나가 다 능동적으로 상위서비스인 golister-watcher@의 execstart를 실행시킬 수 있음

```
//파일명은 : /etc/systemd/system/golister-handler.sh
#!/bin/bash

case "$1" in
    "NAS4-watch-mmd")
        golister 1
        ;;
    "NAS2-priv-watch-pmv")
        golister 2
        ;;
    "NAS3-samba-watch-fancam")
        golister 3
        ;;
    "NAS2-priv-watch-av")
        golister 4
        ;;
    *)
        echo "Unknown instance: $1"
        ;;
esac
```
> 보다시피, @뒤로 이어질 path들의 부속명칭을 인수로 받아 분류별로 명령어를 뱉어내는 단순한 스크립트임



### dpl watcher (윈도우 사이드)
- [[PowerShell Script]]



