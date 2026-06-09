# ADR-0001: 아키텍처 결정 기록을 사용한다

- **상태:** 수락 (Accepted)
- **날짜:** 2026-06-08

## 맥락 (Context)

이 프로젝트의 주요 설계 결정을 추적 가능한 형태로 남길 필요가 있다.

## 결정 (Decision)

모든 구조적 결정은 `docs/adr/ADR-NNNN-<slug>.md` 파일로 기록하고,
[adr/README.md](./README.md) 인덱스에 한 줄을 추가한다.

## 결과 (Consequences)

- 결정의 맥락이 보존되어 신규 기여자가 "왜"를 추적할 수 있다.
- 결정을 바꿀 때는 기존 ADR을 폐기(Superseded)로 표시하고 새 ADR을 추가한다.
