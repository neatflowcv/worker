# ADR 0001: Backlog Item Ordering

## 상태

Accepted

## 문맥

이 저장소는 `worker project backlog ...` 형태의 CLI를 사용합니다.

사용자 입력 관점에서는 다음 인터페이스가 자연스럽습니다.

```bash
worker project backlog move <project> <id> [--after <prev-id>]
```

하지만 `--after <id>`는 사용자 입력 방식일 뿐,
영속 저장 모델로 쓰기에는 불편합니다.

우리가 필요한 것은 다음입니다.

- 프로젝트별 안정 정렬
- 중간 삽입과 이동의 낮은 비용
- `list <project> [--after <id>] [--limit <n>]`와의 자연스러운 결합
- 구현 복잡도의 통제

자세한 비교는 [../backlog-ordering.md](../backlog-ordering.md)에 정리했습니다.

## 결정

`BacklogItem`의 순서는 fractional indexing 기반 `orderKey string`으로 표현합니다.

즉:

- CLI 입력은 `--after <id>`를 사용합니다.
- 내부 저장은 `BacklogItem.orderKey`를 사용합니다.
- 조회 정렬은 `(project_id, order_key ASC)`를 기준으로 합니다.

## 선택하지 않은 방안

### 연속 정수

- 구현은 단순하지만
  중간 삽입과 이동 시 대량 재번호가 필요합니다.

### 간격 정수

- 초기에는 실용적이지만
  반복 삽입 시 재번호가 다시 필요합니다.

### `prev_id` 연결 리스트

- `--after` 의미와는 맞지만
  조회 정렬, 무결성 검증, 복구가 불편합니다.

### LexoRank

- 대규모 시스템에는 강하지만
  현재 단계에서는 fractional indexing보다 구현 부담이 큽니다.

### CRDT sequence identifier

- 동시 협업에는 강하지만
  현재 요구사항에는 과합니다.

## 결과

장점:

- 이동과 중간 삽입 비용이 낮습니다.
- 조회 정렬이 단순합니다.
- CLI와 저장 모델의 책임이 분리됩니다.
- 미래에 bulk key generation이나 rebalance를 붙이기 쉽습니다.

주의점:

- 문자열 비교는 바이트 기준이어야 합니다.
- DB collation이 정렬을 깨지 않도록 확인해야 합니다.
- 특정 구간 재삽입이 반복되면
  장기적으로 rebalance 전략이 필요할 수 있습니다.

## 링크

- 비교 문서: [../backlog-ordering.md](../backlog-ordering.md)
