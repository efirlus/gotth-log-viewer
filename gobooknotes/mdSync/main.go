package main

import (
	"fmt"
	"mdsync/lg"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

func main() {
	src := `
방송통신심의위원회(방심위)가 인터넷 백과사전 [[나무위키에]] 기재된 특정 인플루언서 페이지에 ‘접속차단’을 의결했다. 불법정보가 아닌 권리침해정보 심의로 나무위키에 방심위가 접속차단 의결을 한 건 이번이 처음이다.

방심위는 16일 [[통신심의소위원회]](통신소위)를 열고 인플루언서 A씨와 B씨의 나무위키 페이지가 사생활 침해라며 접속차단을 의결했다. 방심위는 접속차단 전 [[플랫폼]] 관계자의 의견진술을 들을 수 있지만 이번엔 의견진술 없이 통신사(ISP, [[인터넷서비스사업자]])에 URL 차단을 요청한 것으로 확인됐다.

연합뉴스에 따르면 방송 출연 경험이 있는 인플루언서 A씨는 나무위키에 노출된 [[전 연인]]과의 노출 및 스킨십 사진으로 성적 수치심을 느낀다며 방심위에 삭제를 요청했다. 인플루언서 B씨도 나무위키에 본인 동의 없이 2013년부터 2023년까지의 생애와 사진, 본명, 출생, 국적, 신체, 학력, [[수상 경력]]까지 나와 있다고 삭제를 요청했다.

방심위 통신자문특별위원회는 [[인플루언서 A씨]]와 B씨의 민원에 대해 신고인이 원하지 않고 사생활 침해 소지가 있다며 시정 요구가 필요하다는 다수 의견을 냈다. 방심위 역시 이에 근거해 접속차단을 의결한 것으로 알려졌다.

익명의 방심위 관계자는 연합뉴스에 “기존 기조를 바꾼 첫 번째 사례”라며 “해외에 있는 사이트라 개별 삭제 차단 요청을 할 수는 없으나 이렇게 계속 [[의결 및 경고]]를 하고, 시정이 되지 않으면 사례 누적을 확인해 나무위키 전체에 대한 차단도 할 수 있다”고 했다.

출처 : 미디어오늘(https://www.mediatoday.co.kr)
`

	converter := md.NewConverter("", true, nil)

	tempmarkdown, err := converter.ConvertString(src)
	if err != nil {
		lg.Err("마크다운 문법 생성 실패", err)
	}
	fmt.Println(tempmarkdown)

	mdf := replaceBrackets(tempmarkdown)
	fmt.Println(mdf)

}

func replaceBrackets(content string) string {
	// Replace escaped brackets with regular brackets
	content = strings.ReplaceAll(content, `\[\[`, `[[`)
	content = strings.ReplaceAll(content, `\]\]`, `]]`)
	return content
}
