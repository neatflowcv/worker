# Backlog Item Ordering

## 목적

이 문서는 backlog item의 순서를 도메인과 저장소에 어떻게 표현할지 비교하고,
이 저장소에 적합한 기본 방안을 제안합니다.

커맨드에서는 `backlog`라고 부르지만,
도메인에서는 실제 엔터티를 `backlog item`으로 봅니다.

최종 결정은 [ADR: Backlog Item Ordering](adr/0001-backlog-item-ordering.md)에 정리합니다.

## 요구사항

- 프로젝트별 backlog item을 안정적으로 정렬할 수 있어야 합니다.
- `move <project> <id> [--after <prev-id>]`를 자연스럽게 지원해야 합니다.
- 앞, 뒤, 중간 삽입이 자주 발생해도 전체 재번호 비용이 과도하지 않아야 합니다.
- 나중에 `list <project> [--after <id>] [--limit <n>]` 같은 조회와 잘 맞아야 합니다.
- 가능하면 도메인 규칙이 단순해야 합니다.

## 후보 방안

### 1. 연속 정수 위치

예시:

```text
position: 1, 2, 3, 4
```

장점:

- 가장 단순합니다.
- 정렬과 디버깅이 쉽습니다.

단점:

- 중간 삽입이나 이동 시 뒤쪽 항목을 대량 갱신해야 합니다.
- drag and drop이 잦으면 쓰기 비용이 빠르게 커집니다.
- 동시성 충돌에 약합니다.

적합한 경우:

- 항목 수가 매우 작고
- 이동이 드물며
- 구현 단순성이 절대적으로 중요한 경우

### 2. 간격을 둔 정수 위치

예시:

```text
position: 100, 200, 300
```

중간 삽입 시:

```text
250
```

장점:

- 단순한 정수 정렬을 유지하면서
  매번 전체 재번호를 하지 않아도 됩니다.
- 소규모 서비스에서는 실용적입니다.

단점:

- 같은 구간에 삽입이 반복되면 결국 재번호가 필요합니다.
- 벌크 이동 시 새 간격 계산 규칙을 따로 둬야 합니다.

적합한 경우:

- 단일 작성자 중심
- 동시성 요구가 낮음
- 구현 난이도를 낮게 유지하고 싶음

### 3. 연결 리스트 방식

예시:

```text
prev_id: null
prev_id: backlog-1
prev_id: backlog-7
```

장점:

- `--after <id>`와 의미가 직접 맞닿아 있습니다.
- 이동 시 보통 소수 항목만 갱신하면 됩니다.

단점:

- 조회 시 정렬 결과를 바로 얻기 어렵습니다.
- 역참조나 순환 검증이 필요합니다.
- DB 질의만으로 안정 정렬을 만들기 어렵고,
  애플리케이션 레벨 재구성이 필요합니다.
- 손상 복구와 무결성 검사가 까다롭습니다.

적합한 경우:

- 메모리 안에서만 잠깐 쓰는 자료구조
- 영속 저장소 정렬키로는 비추천

### 4. LexoRank 스타일 문자열 랭크

예시:

```text
rank: "0|hzztzz"
rank: "0|i0007n"
```

장점:

- 문자열 정렬만으로 순서를 표현할 수 있습니다.
- 중간 삽입 비용이 낮습니다.
- 대규모 backlog, 빈번한 이동, 다중 서버 환경에서 검증된 방식입니다.

단점:

- 직접 구현하면 규칙이 다소 복잡합니다.
- 장기간 운영 시 rebalance가 필요할 수 있습니다.
- DB collation을 신경 써야 합니다.

적합한 경우:

- 항목 수가 많고
- 이동이 잦고
- 운영 안정성이 중요한 경우

참고:

- Atlassian Jira는 LexoRank를 문자열의 lexicographical order로 저장하며,
  재정렬을 위한 bucket과 rebalance 개념을 둡니다.

### 5. Fractional indexing 문자열 키

예시:

```text
order_key: "a0"
order_key: "a1"
order_key: "a1V"
```

장점:

- 구현이 비교적 단순합니다.
- 문자열 정렬만으로 순서를 표현할 수 있습니다.
- 항목 이동 시 보통 한 항목의 `order_key`만 바꾸면 됩니다.
- 앞, 뒤, 중간 삽입을 자연스럽게 지원합니다.
- `move --after`와 궁합이 좋습니다.

단점:

- 같은 좁은 구간에 삽입이 반복되면 키 길이가 커질 수 있습니다.
- 동시 삽입 충돌을 피하려면 서버 생성 규칙이나 jitter 전략이 필요합니다.
- locale-aware 정렬을 쓰면 잘못 정렬될 수 있습니다.

적합한 경우:

- backlog reorder가 핵심 기능
- 구조는 단순하게 유지하고 싶음
- 미래에 실시간 협업으로 확장할 가능성은 있지만
  지금 당장 CRDT까지는 원하지 않음

### 6. CRDT sequence identifier

예시:

- Logoot
- LSEQ
- RGA 계열

장점:

- 진짜 다중 작성자, 오프라인, peer-to-peer 협업에 강합니다.
- 중앙 서버 없이도 순서 충돌을 병합할 수 있습니다.

단점:

- 지금 요구사항에 비해 과합니다.
- 구현과 디버깅 비용이 큽니다.
- identifier 크기와 병합 규칙이 복잡합니다.

적합한 경우:

- local-first
- 오프라인 편집
- 다중 사용자 동시 reorder
- 중앙 순서 조정자 없이 병합 필요

## 최신 실무 경향

2026-04-10 기준으로 실무형 backlog/list ordering에서는
여전히 문자열 기반 순서 키가 가장 현실적입니다.

특히 많이 보이는 흐름은 다음 둘입니다.

- LexoRank 스타일의 lexicographically sortable string rank
- Fractional indexing 기반 string key

Fractional indexing 쪽은
Figma가 오래전부터 ordered sequence에 사용한 접근으로 공개했고,
이후 David Greenspan의 정리와
Rocicorp의 호환 라이브러리들로 실전 사용성이 더 높아졌습니다.

최근에 보이는 발전은 다음 정도입니다.

- 여러 언어에서 호환 구현 제공
- 여러 키를 한 번에 생성해 길이 증가를 줄이는 API 제공
- 동시 생성 충돌 확률을 줄이기 위한 jitter 확장

반면,
순수 연구 관점에서는 order-maintenance 자료구조와 sequence CRDT 연구가 계속 나오지만,
일반적인 제품 backlog 정렬에는 과한 경우가 많습니다.

## 이 저장소에 대한 추천

기본 추천은 `project_id + order_key` 조합입니다.

예시:

```text
project_id: 01J...
order_key: "a1V"
```

정렬 규칙:

- 같은 프로젝트 안에서 `order_key` 오름차순 정렬

이유:

- 현재 CLI가 `move <project> <id> [--after <prev-id>]` 중심이라
  중간 삽입이 핵심입니다.
- 연결 리스트보다 조회가 쉽습니다.
- 단순 정수보다 이동 비용이 낮습니다.
- LexoRank보다 구현 부담이 낮습니다.
- 나중에 필요하면
  fractional indexing에서 LexoRank 스타일로 진화시켜도
  도메인 인터페이스는 유지하기 쉽습니다.

## 사용자 입력과 저장 표현 분리

이 문서에서 가장 중요한 포인트는
사용자가 말하는 순서와
시스템이 저장하는 순서 표현을 분리하는 것입니다.

사용자는 보통 이렇게 말합니다.

```text
backlog-7을 backlog-3 뒤로 이동
```

그래서 CLI는 다음처럼 두는 것이 자연스럽습니다.

```bash
worker project backlog move <project> <id> [--after <prev-id>]
```

하지만 저장소에 `prev_id`만 영속화하면
조회 정렬, 무결성 검증, 손상 복구가 어려워집니다.

그래서 내부적으로는 각 item에
정렬 전용 값인 `orderKey`를 둡니다.

예시:

```text
A  a0
B  a1
C  a2
```

여기서 `C`를 `A` 뒤로 옮기고 싶으면
사용자 입력은 `after=A`이지만,
내부 저장은 `A`와 `B` 사이에 들어가는 새 `orderKey`를 생성합니다.

예시 결과:

```text
A  a0
C  a0V
B  a1
```

즉:

- `--after <id>`는 사용자 입력 인터페이스
- `orderKey`는 저장과 정렬을 위한 내부 표현

이 문서에서 fractional indexing을 추천하는 이유는
바로 이 `orderKey`를 효율적으로 만들기 좋기 때문입니다.

## 권장 도메인 속성

`BacklogItem`에 다음 속성을 두는 방안을 권장합니다.

```go
type BacklogItem struct {
    id          string
    projectID   string
    title       string
    description string
    status      BacklogItemStatus
    orderKey    string
}
```

설명:

- `id`: 항목 식별자
- `projectID`: 프로젝트 범위
- `status`: `open`, `running`, `blocked`, `done`
- `orderKey`: 프로젝트 내부 정렬 키

중요:

- 순서는 `prev_id`가 아니라 `orderKey`를 기준으로 영속화하는 편이 좋습니다.
- CLI의 `--after <id>`는 입력 인터페이스이고,
  저장 모델은 `orderKey`가 더 적합합니다.

## move 처리 방식 제안

`move <project> <id> [--after <prev-id>]`는 내부적으로 이렇게 처리합니다.

### 맨 앞으로 이동

- `after`가 없으면
  현재 첫 항목의 `orderKey`보다 앞선 새 키를 생성합니다.

### 특정 항목 뒤로 이동

- `after` 항목의 `orderKey`와
  그 다음 항목의 `orderKey` 사이 새 키를 생성합니다.

### 끝으로 이동

- `after`가 마지막 항목이면
  그 뒤에 오는 새 키를 생성합니다.

## 운영 시 주의점

- 문자열 비교는 locale-aware 비교가 아니라 바이트 기준 비교를 사용해야 합니다.
- DB collation이 문자열 순서를 깨지 않도록 확인해야 합니다.
- 한 구간에 재삽입이 과하게 반복되면
  background rebalance를 나중에 추가할 수 있어야 합니다.
- 여러 항목을 한 번에 삽입할 때는
  가능한 경우 여러 키를 균등하게 생성하는 API가 유리합니다.

## 단계별 도입안

### 1단계

- `BacklogItem.orderKey string` 추가
- 프로젝트별 `(project_id, order_key)` 정렬 규칙 도입
- create, move, list를 이 기준으로 구현

### 2단계

- `generateNKeysBetween` 같은 벌크 삽입 전략 도입
- rebalance 관리 커맨드 또는 내부 maintenance 추가 검토

### 3단계

- 실시간 협업이 필요해지면
  jittered fractional indexing 또는
  sequence CRDT 전환 여부 재검토

## 결론

현재 요구사항에는 fractional indexing 기반 `orderKey string`이 가장 적합합니다.

ADR 결정은
[adr/0001-backlog-item-ordering.md](adr/0001-backlog-item-ordering.md)에
기록합니다.

즉:

- CLI 입력: `--after <id>`
- 도메인 저장: `orderKey`
- 조회 정렬: `(project_id, order_key ASC)`

이 조합이 구현 복잡도, 이동 비용, 향후 확장성의 균형이 가장 좋습니다.

## 참고 자료

- [Figma, Realtime editing of ordered sequences](https://www.figma.com/blog/realtime-editing-of-ordered-sequences/)
- [David Greenspan, Implementing Fractional Indexing](https://observablehq.com/@dgreensp/implementing-fractional-indexing)
- [Rocicorp fractional-indexing](https://github.com/rocicorp/fractional-indexing)
- [Rocicorp fracdex Go package](https://pkg.go.dev/roci.dev/fracdex)
- [Atlassian, Troubleshooting LexoRank System Issues](https://support.atlassian.com/jira/kb/troubleshooting-lexorank-system-issues/)
- [Brice Nedelec et al., LSEQ: an Adaptive Structure for Sequences in Distributed
  Collaborative Editing](https://hal.archives-ouvertes.fr/hal-00921633)
