package main

import (
	"fmt"
	"os"
	"strconv"

	conf "golister/module/conf"
	indexing "golister/module/index"
	iomod "golister/module/io"
	lg "golister/module/lg"
	shuffle "golister/module/sort"
)

// TODO: 에러가 나면 각 항목별로 "원래 걸" 찾아나가는 방향으로 빌드해보자, 그 다음엔 Fatal에러를 전부 Error로 낮출 수 있음
// TODO: 재생목록의 모바일버전을 반드시 만들고

func main() {
	lg.NewLogger("goLister")
	lg.Info("======== initiated ========")
	if len(os.Args) != 2 { // golister 에 인수가 하나도 없을 때
		fmt.Println("Usage: golister <number>\n\n1. MMD horizontal\n2. PMV\n3. Fancam\n4. AV")
		return
	}

	choice, err := strconv.Atoi(os.Args[1]) // 인수 확인
	if err != nil {
		lg.Err("error occurs while execution got argument", err)
		return
	}

	// 0. Make some Path related Variables
	Conf := conf.VariableBuilder(choice)

	//1. list
	재생목록 := iomod.BuildMediaList(Conf)

	// 2. directory parser
	폴더, err := indexing.DirectoryLister(Conf.MediaDirectory)
	if err != nil {
		lg.Err("failed listing directory", err)
	}

	// 3. shuffling
	result, forMobile := shuffle.ListRebuilder(재생목록, 폴더, Conf.ModName)

	// 4. file writing
	if err := iomod.CreatePlayList(result, forMobile, Conf.MediaDirectory, Conf.ModName); err != nil {
		lg.Err("writeLines: %s", err)
	}
}
