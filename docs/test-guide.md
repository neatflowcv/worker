# 테스트 가이드

## 목적

이 문서는 이 저장소에서 Go 테스트를 작성할 때 지켜야 하는 최소 규칙을 정리합니다.

## 기본 규칙

- 테스트 파일은 대상 패키지 옆에 `*_test.go`로 둡니다.
- 테스트 패키지명은 반드시 `_test`를 붙입니다.
- 가능하면 외부 공개 동작을 기준으로 검증합니다.
- assertion은 `github.com/stretchr/testify/require`를 사용합니다.
- 테스트에서 `context.Context`가 필요하면 `context.Background()` 대신 `t.Context()`를 사용합니다.

예시:

```go
package flow_test
```

## AAA 형식

모든 테스트는 AAA 형식을 지킵니다.

- Arrange: 테스트 대상과 입력값 준비
- Act: 실제 메서드 호출
- Assert: 결과 검증

예시:

```go
func TestService_CreateProject(t *testing.T) {
 t.Parallel()

 // Arrange
 service := flow.NewService()

 // Act
 project, err := service.CreateProject(t.Context(), "worker", "https://github.com/neatflowcv/worker.git")

 // Assert
 require.NoError(t, err)
 require.NotNil(t, project)
}
```

## 네이밍 규칙

- 테스트 함수명은 `TestXxx`
- 하나의 테스트는 한 가지 동작만 검증합니다.
- 테스트 이름만 보고 무엇을 보장하는지 알 수 있어야 합니다.

## 실행 명령

```bash
go test ./...
go test ./internal/app/flow
```

## 실행 원칙

- 테스트 검증 시에는 반드시 `go test ./...`처럼 전체 테스트를 먼저 실행합니다.
- 특정 패키지 테스트 실행은 빠른 확인용으로만 사용하고, 작업 마무리 전에는 전체 테스트 결과를 확인합니다.
