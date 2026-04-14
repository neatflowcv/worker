# GORM v2 가이드

## 목적

이 문서는 PostgreSQL 기반 저장소 구현을 추가할 때 참고할 수 있도록
GORM v2의 핵심 사용법을 정리합니다.

기준 문서는 GORM 공식 문서이며,
마지막 확인 일자는 2026-04-14입니다.

## 공식 문서

- 메인 가이드: <https://gorm.io/docs/>
- Generics API: <https://gorm.io/docs/the_generics_way.html>
- PostgreSQL 연결: <https://gorm.io/docs/connecting_to_the_database.html>
- 모델 선언: <https://gorm.io/docs/models.html>
- 규칙과 네이밍: <https://gorm.io/docs/conventions.html>
- 조회: <https://gorm.io/docs/query.html>
- 생성: <https://gorm.io/docs/create.html>
- 수정: <https://gorm.io/docs/update.html>
- 삭제: <https://gorm.io/docs/delete.html>
- 트랜잭션: <https://gorm.io/docs/transactions.html>
- 마이그레이션: <https://gorm.io/docs/migration.html>
- 에러 처리: <https://gorm.io/docs/error_handling.html>
- Context 사용: <https://gorm.io/docs/context.html>

## 설치

GORM 공식 문서 기준으로 generics API는
`gorm` `v1.30.0` 이상에서 사용할 수 있습니다.

```bash
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

## PostgreSQL 연결

GORM v2에서 PostgreSQL은 `gorm.io/driver/postgres` 드라이버를 사용합니다.

```go
package postgresql

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func Open(dsn string) (*gorm.DB, error) {
    return gorm.Open(
        postgres.Open(dsn),
        &gorm.Config{
            TranslateError: true,
        },
    )
}
```

예시 DSN:

```text
host=localhost user=worker password=secret dbname=worker
port=5432 sslmode=disable TimeZone=Asia/Seoul
```

`TranslateError: true`를 켜면
중복 키, 외래 키 위반 같은 데이터베이스별 오류를
GORM 공통 오류로 다루기 쉬워집니다.

## 연결 풀 설정

`gorm.DB`를 연 뒤에는 내부 `*sql.DB`를 꺼내 연결 풀을 설정할 수 있습니다.

```go
package postgresql

import "time"

func ConfigurePool(db *gorm.DB) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }

    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(20)
    sqlDB.SetConnMaxLifetime(time.Hour)

    return nil
}
```

애플리케이션 종료 시에는 `sqlDB.Close()`를 호출합니다.

## 모델 선언

GORM 모델은 일반 Go struct로 정의합니다.
기본적으로 `ID` 필드는 primary key,
`CreatedAt`, `UpdatedAt`은 타임스탬프로 인식합니다.

```go
package postgresql

import (
    "database/sql"
    "time"
)

type ProjectModel struct {
    ID            string         `gorm:"primaryKey;type=text"`
    Name          string         `gorm:"uniqueIndex;not null"`
    RepositoryURL string         `gorm:"not null"`
    AuthUsername  sql.NullString
    AuthPassword  sql.NullString
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

기본 규칙:

- struct 이름은 기본적으로 `snake_case` 복수형 테이블명으로 변환됩니다.
- 필드명은 기본적으로 `snake_case` 컬럼명으로 변환됩니다.
- DB 레이어의 nullable 값은 `sql.NullString` 같은 `database/sql` 타입으로 표현합니다.

이 저장소에서는 DB 레이어와 도메인 레이어를 분리하므로
GORM 모델에서는 `sql.NullString`을 사용하고,
도메인으로 올릴 때 포인터나 도메인 타입으로 변환하는 편이 낫습니다.

예시:

```go
func toNullString(v *string) sql.NullString {
    if v == nil {
        return sql.NullString{}
    }

    return sql.NullString{
        String: *v,
        Valid:  true,
    }
}

func toStringPtr(v sql.NullString) *string {
    if !v.Valid {
        return nil
    }

    s := v.String

    return &s
}
```

테이블명을 고정하고 싶으면 `TableName()`을 구현합니다.

```go
func (ProjectModel) TableName() string {
    return "projects"
}
```

## 이 저장소에서의 권장 구조

이 저장소의 [docs/go-guide.md](docs/go-guide.md)에 따라
인터페이스와 구현체는 다른 패키지에 둡니다.

권장 구조:

```text
internal/pkg/repository/
  project.go
  backlog_item.go

internal/pkg/postgresql/
  database.go
  migrate.go
  models.go
  project_repository.go
  backlog_item_repository.go
```

권장 원칙:

- `internal/pkg/repository`에는 인터페이스만 둡니다.
- `internal/pkg/postgresql`에는 GORM 구현체만 둡니다.
- 도메인 모델과 GORM 모델은 분리합니다.
- GORM 태그는 구현체 패키지 내부에만 둡니다.

## Context 사용

저장소 메서드는 이미 `context.Context`를 받으므로,
GORM 호출도 같은 context를 넘겨야 합니다.
새 코드에서는 generics API를 우선 사용하는 편이 낫습니다.

```go
func (r *ProjectRepository) GetProjectByName(
    ctx context.Context,
    name string,
) (*domain.Project, error) {
    model, err := gorm.G[ProjectModel](r.db).
        Where("name = ?", name).
        First(ctx)
    if err != nil {
        return nil, err
    }

    return toDomainProject(model), nil
}
```

## 생성

단건 생성은 `Create`를 사용합니다.

```go
model := ProjectModel{
    ID:            project.ID(),
    Name:          project.Name(),
    RepositoryURL: project.RepositoryURL(),
}

if err := gorm.G[ProjectModel](db).Create(ctx, &model); err != nil {
    return err
}
```

여러 건 생성도 가능하지만,
현재 저장소 구현 요구에는 단건 위주가 더 자연스럽습니다.

## 조회

단건 조회는 보통 `First` 또는 `Take`를 사용합니다.

```go
model, err := gorm.G[ProjectModel](db).
    Where("name = ?", name).
    First(ctx)
```

여러 건 조회는 `Find`를 사용합니다.

```go
models, err := gorm.G[ProjectModel](db).
    Order("name asc").
    Find(ctx)
```

주의:

- `First`, `Last`, `Take`는 없을 때 `gorm.ErrRecordNotFound`를 반환합니다.
- 단건 조회에서 `Find`를 쓰면 없을 때 에러가 나지 않아 의미가 흐려질 수 있습니다.
- 정렬이 필요한 목록은 반드시 `Order(...)`를 명시합니다.

## 수정

부분 수정은 `Updates`를 우선 사용합니다.

```go
rows, err := gorm.G[BacklogItemModel](db).
    Where("id = ?", item.ID()).
    Updates(ctx, map[string]any{
        "title":       item.Title(),
        "description": item.Description(),
        "status":      item.Status(),
        "order_key":   item.OrderKey(),
    })

if err != nil {
    return err
}

if rows == 0 {
    return repository.ErrBacklogItemNotFound
}
```

전체 수정을 하고 싶다면 `Select("*")` 또는
업데이트할 필드를 명시하는 `Select("...")`를 사용할 수 있습니다.

전체 필드 반영:

```go
rows, err := gorm.G[BacklogItemModel](db).
    Where("id = ?", item.ID()).
    Select("*").
    Updates(ctx, BacklogItemModel{
        ID:          item.ID(),
        ProjectID:   item.ProjectID(),
        Title:       item.Title(),
        Description: item.Description(),
        Status:      string(item.Status()),
        OrderKey:    item.OrderKey(),
    })

if err != nil {
    return err
}

if rows == 0 {
    return repository.ErrBacklogItemNotFound
}
```

업데이트 대상 필드만 명시:

```go
rows, err := gorm.G[BacklogItemModel](db).
    Where("id = ?", item.ID()).
    Select("project_id", "title", "description", "status", "order_key").
    Updates(ctx, BacklogItemModel{
        ProjectID:   item.ProjectID(),
        Title:       item.Title(),
        Description: item.Description(),
        Status:      string(item.Status()),
        OrderKey:    item.OrderKey(),
    })

if err != nil {
    return err
}

if rows == 0 {
    return repository.ErrBacklogItemNotFound
}
```

주의:

- `Updates(struct)`는 zero value를 기본적으로 제외합니다.
- 빈 문자열, `false`, `0`까지 반영해야 하면 `map[string]any`가 안전합니다.
- 전체 필드를 반영하려면 `Select("*").Updates(ctx, struct)`를 사용합니다.
- 기본 키나 감사 컬럼까지 건드리기 싫으면 `Select("field1", ...)`가 더 안전합니다.
- generics API에는 `Save`가 없습니다.
- 저장소 구현에서는 `Create`와 `Updates`를 분리해서 쓰는 편이 안전합니다.

## 삭제

삭제는 `Delete`를 사용합니다.

```go
err := gorm.G[ProjectModel](db).
    Where("id = ?", id).
    Delete(ctx)
```

현재 저장소 인터페이스에는 삭제 메서드가 없으므로
필요해질 때 추가하는 편이 낫습니다.

## 트랜잭션

여러 SQL을 하나의 원자 작업으로 묶어야 하면 `Transaction`을 사용합니다.

```go
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := gorm.G[ProjectModel](tx).Create(ctx, &project); err != nil {
        return err
    }

    if err := gorm.G[BacklogItemModel](tx).
        Create(ctx, &backlogItem); err != nil {
        return err
    }

    return nil
})
```

규칙:

- 트랜잭션 안에서는 반드시 `tx`를 계속 사용합니다.
- 일부만 성공하면 안 되는 작업에만 트랜잭션을 씁니다.
- 저장소 경계에서 트랜잭션을 시작할지,
  서비스에서 여러 저장소를 묶을지 먼저 정해야 합니다.

## 마이그레이션

GORM은 `AutoMigrate`를 제공합니다.

```go
func Migrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &ProjectModel{},
        &BacklogItemModel{},
    )
}
```

공식 문서 기준으로 `AutoMigrate`는
테이블, 누락된 컬럼, 인덱스, 제약조건을 추가하거나 일부 타입 변경을 수행하지만,
사용하지 않는 컬럼을 삭제하지는 않습니다.

권장 방향:

- 초기 단계에서는 `AutoMigrate`로 시작할 수 있습니다.
- 운영 환경이 커지면 명시적 migration 도구 도입을 검토합니다.
- 애플리케이션 시작 시 자동 실행할지,
  별도 커맨드로 분리할지 명확히 정합니다.

## 에러 처리

기본 패턴:

```go
model, err := gorm.G[ProjectModel](db).
    Where("name = ?", name).
    First(ctx)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, repository.ErrProjectNotFound
    }

    if errors.Is(err, gorm.ErrDuplicatedKey) {
        return nil, ErrProjectAlreadyExists
    }

    return nil, fmt.Errorf("get project by name: %w", err)
}
```

권장 규칙:

- `gorm.ErrRecordNotFound`는 저장소 인터페이스의 도메인 에러로 변환합니다.
- 데이터베이스 드라이버 에러를 서비스 계층으로 직접 노출하지 않습니다.
- `TranslateError: true`를 켠 뒤 `gorm.ErrDuplicatedKey` 같은 공통 오류를 우선 사용합니다.

## 이 저장소에 맞는 구현 팁

`ProjectRepository` 구현 팁:

- `name`에 unique index를 둡니다.
- `GetProjectByName`는
  `gorm.G[ProjectModel](db).Where("name = ?", name).First(ctx)`로 구현합니다.
- `ListProjects`는 `Order("name asc")` 같은 명시 정렬을 둡니다.

`BacklogItemRepository` 구현 팁:

- `(project_id, order_key)` 복합 인덱스를 둡니다.
- 목록 조회는 `Where("project_id = ?", projectID).Order("order_key asc")`로 구현합니다.
- `afterID`가 있으면 먼저 해당 row를 읽고,
  같은 프로젝트인지 확인한 뒤 `order_key > ?` 조건을 추가합니다.
- `Updates(ctx, map[string]any{...})`를 사용해 zero value 누락을 피합니다.

## 최소 예시

```go
package postgresql

import (
    "context"
    "errors"
    "fmt"

    "github.com/neatflowcv/worker/internal/pkg/domain"
    "github.com/neatflowcv/worker/internal/pkg/repository"
    "gorm.io/gorm"
)

type ProjectRepository struct {
    db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
    return &ProjectRepository{db: db}
}

func (r *ProjectRepository) GetProjectByName(
    ctx context.Context,
    name string,
) (*domain.Project, error) {
    model, err := gorm.G[ProjectModel](r.db).
        Where("name = ?", name).
        First(ctx)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, repository.ErrProjectNotFound
        }

        return nil, fmt.Errorf("query project by name: %w", err)
    }

    return toDomainProject(model), nil
}
```

## 참고

- 저장소 인터페이스: [internal/pkg/repository/project.go](../internal/pkg/repository/project.go)
- 저장소 인터페이스: [internal/pkg/repository/backlog_item.go](../internal/pkg/repository/backlog_item.go)
- Go 규칙: [docs/go-guide.md](go-guide.md)
- 테스트 규칙: [docs/test-guide.md](test-guide.md)
