// Package schema defines the standardized documentation pattern that docs-cli
// produces and validates. It is the single source of truth: the scaffolder,
// the agent prompt builder, the validator, and the renderer all read from it.
//
// The pattern is intentionally fixed so that generated docs are uniform across
// projects and can be ingested by a web service without per-project mapping:
// every document declares stable frontmatter keys and a known, ordered set of
// chapter ids.
package schema

// SchemaVersion is the version of the standardized documentation pattern.
// Bump this when the document set or required frontmatter changes.
const SchemaVersion = "1"

// Status values used in document frontmatter.
const (
	StatusDraft     = "draft"     // scaffolded template, not yet filled
	StatusGenerated = "generated" // filled by an AI agent, awaiting review
	StatusReviewed  = "reviewed"  // human-reviewed and accepted
)

// Section groups documents into a reading flow, mirroring a requirements →
// architecture → contract → operations → decisions → reference progression.
type Section struct {
	ID    string
	Title string
	// Order controls section ordering in the index and frontmatter.
	Order int
	// Blurb is a one-line description of the section's purpose.
	Blurb string
}

// Chapter is a required heading inside a document. The Guidance is shown to
// AI agents (what to write) and embedded as a scaffold placeholder.
type Chapter struct {
	// ID is a stable slug used by the web service to address a chapter.
	ID string
	// Heading is the rendered Markdown H2 text.
	Heading string
	// Guidance tells an agent what belongs in this chapter.
	Guidance string
}

// Document describes a single standardized document.
type Document struct {
	// ID is the stable slug; the file is written as "<ID>.md".
	ID string
	// Title is the document H1 and frontmatter title.
	Title string
	// Section groups the document; must match a Section.ID.
	Section string
	// Order is the position within its section.
	Order int
	// Purpose is a one-line summary used in the index and frontmatter.
	Purpose string
	// Audience lists who primarily reads this document.
	Audience []string
	// Required marks documents that must exist for a valid docs set.
	Required bool
	// Dir, when set, places the file under a subdirectory and signals that
	// the document is the index of a collection (e.g. adr/, diagrams/).
	Dir string
	// Chapters are the ordered, required H2 sections.
	Chapters []Chapter
}

// Schema is the full standardized documentation pattern.
type Schema struct {
	Version  string
	Sections []Section
	Docs     []Document
}

// FileName returns the path (relative to the docs root) for the document.
func (d Document) FileName() string {
	if d.Dir != "" {
		return d.Dir + "/README.md"
	}
	return d.ID + ".md"
}

// Standard returns the canonical schema (version 1).
func Standard() Schema {
	return Schema{
		Version:  SchemaVersion,
		Sections: standardSections,
		Docs:     standardDocs,
	}
}

// Section returns the section with the given id, and whether it was found.
func (s Schema) SectionByID(id string) (Section, bool) {
	for _, sec := range s.Sections {
		if sec.ID == id {
			return sec, true
		}
	}
	return Section{}, false
}

// Doc returns the document with the given id, and whether it was found.
func (s Schema) Doc(id string) (Document, bool) {
	for _, d := range s.Docs {
		if d.ID == id {
			return d, true
		}
	}
	return Document{}, false
}

// DocsInSection returns the documents that belong to a section, in order.
func (s Schema) DocsInSection(sectionID string) []Document {
	var out []Document
	for _, d := range s.Docs {
		if d.Section == sectionID {
			out = append(out, d)
		}
	}
	return out
}

var standardSections = []Section{
	{ID: "overview", Order: 1, Title: "개요 (Overview)", Blurb: "이 프로젝트가 무엇이고 왜 존재하는가"},
	{ID: "architecture", Order: 2, Title: "아키텍처 (Architecture)", Blurb: "구조·모듈·데이터가 어떻게 짜여 있는가"},
	{ID: "interfaces", Order: 3, Title: "계약 (Interfaces)", Blurb: "구현이 그대로 따르는 인터페이스 — API·CLI·설정"},
	{ID: "operations", Order: 4, Title: "운영 (Operations)", Blurb: "어떻게 빌드·배포·테스트·보안 점검하는가"},
	{ID: "decisions", Order: 5, Title: "결정 (Decisions)", Blurb: "왜 이렇게 설계했는가 — 아키텍처 결정 기록"},
	{ID: "reference", Order: 6, Title: "참고 (Reference)", Blurb: "로드맵·추적성 등 횡단 참고 자료"},
}

var standardDocs = []Document{
	{
		ID: "overview", Title: "프로젝트 개요", Section: "overview", Order: 1, Required: true,
		Purpose:  "프로젝트의 목적·문제·사용자·가치를 한눈에",
		Audience: []string{"newcomer", "decision-maker"},
		Chapters: []Chapter{
			{ID: "tldr", Heading: "TL;DR", Guidance: "3~5개 불릿으로 정체성·핵심 가치·차별점을 요약한다."},
			{ID: "problem", Heading: "문제 정의", Guidance: "이 프로젝트가 해결하는 문제와, 해결되지 않으면 생기는 고통을 서술한다."},
			{ID: "audience", Heading: "사용자와 이해관계자", Guidance: "주 사용자·운영자·의사결정자 등 누가 왜 쓰는지 표로 정리한다."},
			{ID: "value", Heading: "핵심 가치와 비-목표", Guidance: "제공하는 가치와, 의도적으로 하지 않는 것(non-goals)을 명시한다."},
			{ID: "stack", Heading: "기술 스택 요약", Guidance: "언어·런타임·핵심 의존성·배포 형태를 표로 요약한다."},
		},
	},
	{
		ID: "glossary", Title: "용어집", Section: "overview", Order: 2, Required: true,
		Purpose:  "표준 용어 + 프로젝트 특화 용어 정의",
		Audience: []string{"newcomer", "integrator"},
		Chapters: []Chapter{
			{ID: "domain", Heading: "도메인 용어", Guidance: "프로젝트 도메인에서 쓰는 명사·약어를 표로 정의한다."},
			{ID: "technical", Heading: "기술 용어", Guidance: "코드/아키텍처에서 반복되는 기술 용어를 정의한다."},
		},
	},
	{
		ID: "architecture", Title: "아키텍처 개요", Section: "architecture", Order: 1, Required: true,
		Purpose:  "시스템 컨텍스트·컨테이너·컴포넌트 뷰",
		Audience: []string{"architect", "newcomer"},
		Chapters: []Chapter{
			{ID: "context", Heading: "시스템 컨텍스트", Guidance: "외부 행위자·시스템과의 경계를 Mermaid 다이어그램과 함께 설명한다."},
			{ID: "containers", Heading: "컨테이너 / 런타임 단위", Guidance: "프로세스·서비스·바이너리 등 배포 단위와 통신 방식을 설명한다."},
			{ID: "quality", Heading: "품질 속성", Guidance: "성능·가용성·보안·확장성 등 아키텍처가 보장해야 하는 품질을 기술한다."},
			{ID: "crosscutting", Heading: "횡단 관심사", Guidance: "로깅·설정·인증·에러처리 등 모듈을 가로지르는 관심사를 정리한다."},
		},
	},
	{
		ID: "components", Title: "컴포넌트 설계", Section: "architecture", Order: 2, Required: true,
		Purpose:  "모듈 경계·책임·의존 규칙",
		Audience: []string{"architect", "contributor"},
		Chapters: []Chapter{
			{ID: "modules", Heading: "모듈 지도", Guidance: "주요 패키지/모듈과 그 책임을 표로 정리하고 디렉터리 트리를 포함한다."},
			{ID: "dependencies", Heading: "의존 규칙", Guidance: "허용/금지된 의존 방향과 그 이유를 Mermaid 의존 그래프로 보인다."},
			{ID: "boundaries", Heading: "경계와 확장점", Guidance: "인터페이스·어댑터·플러그인 등 확장 지점을 설명한다."},
		},
	},
	{
		ID: "data-model", Title: "데이터 모델", Section: "architecture", Order: 3, Required: false,
		Purpose:  "엔티티·관계·저장소 매핑",
		Audience: []string{"contributor", "integrator"},
		Chapters: []Chapter{
			{ID: "entities", Heading: "엔티티", Guidance: "핵심 엔티티/테이블과 필드를 표로 정의한다."},
			{ID: "relationships", Heading: "관계 (ER)", Guidance: "엔티티 간 관계를 Mermaid ER 다이어그램으로 표현한다."},
			{ID: "lifecycle", Heading: "수명주기·정합성", Guidance: "생성·갱신·삭제 규칙과 불변식을 기술한다."},
		},
	},
	{
		ID: "diagrams", Title: "다이어그램", Section: "architecture", Order: 4, Required: false, Dir: "diagrams",
		Purpose:  "시퀀스·상태·플로우 다이어그램 모음",
		Audience: []string{"architect", "contributor"},
		Chapters: []Chapter{
			{ID: "index", Heading: "다이어그램 목록", Guidance: "개별 Mermaid 다이어그램 파일을 한 줄씩 링크하고 무엇을 보여주는지 적는다."},
		},
	},
	{
		ID: "api", Title: "API / 인터페이스 계약", Section: "interfaces", Order: 1, Required: false,
		Purpose:  "외부에 노출되는 인터페이스의 계약",
		Audience: []string{"integrator"},
		Chapters: []Chapter{
			{ID: "surface", Heading: "인터페이스 표면", Guidance: "HTTP 엔드포인트·gRPC·라이브러리 공개 API 등 노출 표면을 카탈로그화한다."},
			{ID: "contracts", Heading: "요청·응답 계약", Guidance: "대표 호출의 입력·출력·에러 코드를 예시와 함께 기술한다."},
			{ID: "versioning", Heading: "호환성·버저닝", Guidance: "하위호환 정책과 변경 절차를 설명한다."},
		},
	},
	{
		ID: "cli-reference", Title: "CLI 레퍼런스", Section: "interfaces", Order: 2, Required: false,
		Purpose:  "명령·플래그·종료 코드",
		Audience: []string{"operator", "integrator"},
		Chapters: []Chapter{
			{ID: "commands", Heading: "명령 카탈로그", Guidance: "모든 명령/서브명령과 목적을 표로 정리한다."},
			{ID: "flags", Heading: "전역 플래그", Guidance: "공통 플래그와 환경 변수를 정리한다."},
			{ID: "exit-codes", Heading: "종료 코드", Guidance: "각 종료 코드의 의미를 표로 정의한다."},
		},
	},
	{
		ID: "config-reference", Title: "설정 레퍼런스", Section: "interfaces", Order: 3, Required: false,
		Purpose:  "설정 키·기본값·검증 규칙",
		Audience: []string{"operator"},
		Chapters: []Chapter{
			{ID: "discovery", Heading: "설정 탐색·우선순위", Guidance: "설정 파일 위치, 환경 변수, 플래그 간 우선순위를 설명한다."},
			{ID: "keys", Heading: "설정 키", Guidance: "모든 설정 키를 타입·기본값·설명과 함께 표로 정의한다."},
			{ID: "examples", Heading: "예시 구성", Guidance: "대표 시나리오별 완전한 설정 예시를 보인다."},
		},
	},
	{
		ID: "build-and-release", Title: "빌드 · 릴리스", Section: "operations", Order: 1, Required: true,
		Purpose:  "빌드·버전 관리·릴리스 파이프라인",
		Audience: []string{"contributor", "operator"},
		Chapters: []Chapter{
			{ID: "build", Heading: "빌드", Guidance: "로컬 빌드 명령과 산출물 위치, 크로스 컴파일 대상을 정리한다."},
			{ID: "versioning", Heading: "버전 관리", Guidance: "SemVer 정책, 버전 주입 방식, 태그 규칙을 설명한다."},
			{ID: "release", Heading: "릴리스 파이프라인", Guidance: "태그→빌드→서명/체크섬→배포 단계를 순서대로 기술한다."},
		},
	},
	{
		ID: "deployment", Title: "배포", Section: "operations", Order: 2, Required: false,
		Purpose:  "배포 형상·환경·롤백",
		Audience: []string{"operator"},
		Chapters: []Chapter{
			{ID: "targets", Heading: "배포 대상", Guidance: "지원 플랫폼/환경과 산출물(바이너리·컨테이너·차트)을 정리한다."},
			{ID: "procedure", Heading: "배포 절차", Guidance: "설치·업그레이드 단계와 검증 방법을 기술한다."},
			{ID: "rollback", Heading: "롤백·복구", Guidance: "실패 시 롤백 절차와 백업/복구 전략을 설명한다."},
		},
	},
	{
		ID: "development", Title: "개발 가이드", Section: "operations", Order: 3, Required: true,
		Purpose:  "개발 환경·작업 루프·기여 규칙",
		Audience: []string{"contributor"},
		Chapters: []Chapter{
			{ID: "setup", Heading: "개발 환경 준비", Guidance: "필수 도구·버전·셋업 명령을 정리한다."},
			{ID: "workflow", Heading: "작업 루프", Guidance: "브랜치·커밋·리뷰·검증의 표준 루프를 기술한다."},
			{ID: "conventions", Heading: "코드 규약", Guidance: "포맷·네이밍·린트 규칙과 강제 장치(pre-commit, CI)를 설명한다."},
		},
	},
	{
		ID: "testing", Title: "테스트 전략", Section: "operations", Order: 4, Required: true,
		Purpose:  "테스트 계층·커버리지·실행",
		Audience: []string{"contributor"},
		Chapters: []Chapter{
			{ID: "layers", Heading: "테스트 계층", Guidance: "단위·통합·e2e 등 계층과 각 계층의 책임을 정리한다."},
			{ID: "running", Heading: "실행 방법", Guidance: "테스트 실행 명령과 CI 게이트를 설명한다."},
			{ID: "coverage", Heading: "커버리지·품질 게이트", Guidance: "커버리지 목표와 통과 기준을 기술한다."},
		},
	},
	{
		ID: "security", Title: "보안 고려사항", Section: "operations", Order: 5, Required: true,
		Purpose:  "위협 표면·비밀 관리·공급망",
		Audience: []string{"security-reviewer", "operator"},
		Chapters: []Chapter{
			{ID: "surface", Heading: "위협 표면", Guidance: "신뢰 경계와 주요 위협(STRIDE 관점)을 정리한다."},
			{ID: "secrets", Heading: "비밀·자격증명 관리", Guidance: "비밀이 어떻게 저장·주입·회전되는지 설명한다."},
			{ID: "supply-chain", Heading: "공급망·의존성", Guidance: "의존성 점검·SBOM·서명 등 공급망 방어를 기술한다."},
		},
	},
	{
		ID: "adr", Title: "아키텍처 결정 기록 (ADR)", Section: "decisions", Order: 1, Required: true, Dir: "adr",
		Purpose:  "주요 설계 결정의 맥락·결정·결과",
		Audience: []string{"architect", "contributor"},
		Chapters: []Chapter{
			{ID: "index", Heading: "결정 목록", Guidance: "개별 ADR 파일(ADR-001.md…)을 표로 인덱싱하고 상태(제안/수락/폐기)를 적는다."},
			{ID: "convention", Heading: "ADR 작성 규약", Guidance: "ADR 번호·상태·템플릿(맥락·결정·결과) 규약을 설명한다."},
		},
	},
	{
		ID: "roadmap", Title: "로드맵", Section: "reference", Order: 1, Required: false,
		Purpose:  "마일스톤·버전 계획",
		Audience: []string{"decision-maker", "contributor"},
		Chapters: []Chapter{
			{ID: "milestones", Heading: "마일스톤", Guidance: "버전별 범위를 표로 정리한다 (current/next/later)."},
			{ID: "now", Heading: "현재 집중", Guidance: "지금 진행 중인 작업과 다음 우선순위를 적는다."},
		},
	},
	{
		ID: "traceability", Title: "추적성 · 문서 커버리지", Section: "reference", Order: 2, Required: true,
		Purpose:  "문서 간 정합성과 커버리지 검증",
		Audience: []string{"maintainer"},
		Chapters: []Chapter{
			{ID: "matrix", Heading: "커버리지 매트릭스", Guidance: "표준 스키마의 각 문서가 작성/생성/검토 중 어느 상태인지 표로 보인다."},
			{ID: "open-items", Heading: "열린 항목 등록부", Guidance: "아직 비어 있거나 미정인 항목을 숨기지 말고 명시한다."},
		},
	},
}
