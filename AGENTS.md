# Repository Guidelines

## 프로젝트 구조

- `cmd/`: 실행 진입점. 하위에는 각 실행 바이너리 디렉터리를 두며 `main` 패키지만 두고 얇게 유지합니다.
- `internal/app/flow/`: 애플리케이션 서비스
- `internal/pkg/domain/`: 도메인 모델. 외부 라이브러리를 사용하지 않고 스탠다드 라이브러리만 허용합니다.
- `docs/`: 작업 규칙 문서

## 상세 규칙

- 테스트 가이드: [docs/test-guide.md](docs/test-guide.md)
  - 테스트를 작성하거나 기존 테스트를 변경할 때 반드시 함께 확인합니다.
- 커밋 컨벤션: [docs/commit-convention.md](docs/commit-convention.md)
  - 커밋 메시지를 작성할 때 반드시 확인합니다.
- 마크다운 가이드: [docs/markdown-guide.md](docs/markdown-guide.md)
  - 마크다운 문서를 작성하거나 수정할 때 반드시 확인합니다.
- Go 가이드: [docs/go-guide.md](docs/go-guide.md)
  - Go 코드를 작성하거나 수정할 때 반드시 확인합니다.
  - Go 코드 변경 작업은 완료 보고 전에 가이드에 적힌 모든 명령을 실제로 순서대로 실행하고, 통과 여부를 확인해야 합니다.
  - 위 점검을 실행하지 않았거나 실패한 상태에서는 작업을 완료로 판단하지 않습니다.

## 문서 링크 규칙

- 문서 링크는 저장소 기준 상대 경로를 사용합니다.
- 사용자 홈 디렉토리나 로컬 절대경로는 문서에 남기지 않습니다.
