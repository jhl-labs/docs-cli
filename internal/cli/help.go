package cli

import (
	"fmt"
	"io"
)

func printHelp(w io.Writer) {
	fmt.Fprint(w, `docs-cli — 표준화된 프로젝트 문서를 생성·검증·렌더링하는 도구

사용법:
  docs-cli <command> [path] [flags]

명령:
  init        표준 스키마에 따라 docs/ 트리를 스캐폴딩한다
  generate    프로젝트를 리버스 엔지니어링하여 문서를 채운다(에이전트 또는 프롬프트)
  validate    docs/ 가 표준 스키마를 따르는지 검증한다
  render      문서를 HTML/XML로 렌더링한다
  doctor      환경과 사용 가능한 AI 에이전트를 점검한다
  skill       AI 에이전트용 SKILL.md 를 생성한다 (= --generate-skill)
  version     버전을 출력한다

전역:
  -h, --help       도움말
  -v, --version    버전

예시:
  docs-cli init .
  docs-cli generate . --agent claude
  docs-cli generate . --dry-run --doc architecture --doc components
  docs-cli validate . --strict
  docs-cli render . --format html --format xml
  docs-cli --generate-skill --output skill.md

종료 코드: 0 성공 · 1 검증 실패 · 2 사용법 오류 · 3 에이전트 실패 · 4 환경 문제
`)
}

func printInitHelp(w io.Writer) {
	fmt.Fprint(w, `docs-cli init [path] — 표준 docs/ 트리를 스캐폴딩

flags:
  --output-dir <dir>   문서 루트 (기본: docs)
  --lang <lang>        주 언어 강제 지정 (기본: auto, 자동 감지)
  --force              기존 파일 덮어쓰기

path 는 분석할 프로젝트 경로(기본: .)이며 언어 자동 감지에 사용된다.
`)
}

func printGenerateHelp(w io.Writer) {
	fmt.Fprint(w, `docs-cli generate [path] — 문서를 채우기 위한 프롬프트 생성/에이전트 실행

flags:
  --agent <name>       claude|codex|gemini|opencode|none (기본: none)
  --doc <id>           특정 문서만 대상 (반복 가능)
  --output-dir <dir>   문서 루트 (기본: docs)
  --prompts-dir <dir>  프롬프트 저장 경로 (기본: .docs-cli/prompts)
  --lang <lang>        주 언어 강제 지정 (기본: auto)
  --dry-run            에이전트를 실행하지 않고 프롬프트만 작성
  --print-prompt       프롬프트를 stdout 에도 출력

--agent none(기본)이면 프롬프트 파일만 생성한다. 에이전트를 지정하면
각 문서를 비대화형으로 실행해 결과를 docs/<id>.md 에 기록한다.
`)
}

func printValidateHelp(w io.Writer) {
	fmt.Fprint(w, `docs-cli validate [path] — 표준 스키마 정합성 검증

flags:
  --input-dir <dir>    문서 루트 (기본: docs)
  --strict             경고(warn)도 실패로 처리

오류가 1개 이상이면 종료 코드 1 을 반환한다.
`)
}

func printRenderHelp(w io.Writer) {
	fmt.Fprint(w, `docs-cli render [path] — 문서를 HTML/XML 로 렌더링

flags:
  --input-dir <dir>    문서 루트 (기본: docs)
  --output-dir <dir>   산출물 루트 (기본: 포맷별 <input>/_site, <input>/_xml)
  --format <fmt>       html|xml (반복 가능, 기본: html)
`)
}

func printDoctorHelp(w io.Writer) {
	fmt.Fprint(w, `docs-cli doctor — 환경 점검

flags:
  --input-dir <dir>    검증할 문서 루트 (기본: docs)
`)
}

func printSkillHelp(w io.Writer) {
	fmt.Fprint(w, `docs-cli skill — AI 에이전트용 SKILL.md 생성

flags:
  --output <path>      출력 경로 (기본: skill.md, '-' 이면 stdout)
  --name <slug>        스킬 이름 (기본: standardizing-project-docs)
  --binary <name>      예시에 쓸 CLI 이름 (기본: docs-cli)
`)
}
