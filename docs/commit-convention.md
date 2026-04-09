# 커밋 컨벤션 조사 정리

## 결론

이 저장소는 **Conventional Commits**를 기본 규칙으로 채택하는 것이 가장 적절합니다.  
이유는 형식이 단순하고, 변경 의도를 빠르게 파악할 수 있으며, 이후 changelog 자동화나 릴리스 규칙과도 잘 맞기 때문입니다.

추가로 제목/본문 길이와 문체는 Git 공식 문서의 권장사항을 따르는 것이 좋습니다.

## 권장 형식

기본 형식:

```text
<type>(<scope>): <subject>
```

예시:

```text
feat(flow): add project creation service
fix(domain): return repository URL consistently
test(flow): cover ULID generation in CreateProject
docs: add repository guidelines
```

본문이 필요하면 아래 형식을 사용합니다.

```text
<type>(<scope>): <subject>

<body>
```

## 이 저장소에 권장하는 type

- `feat`: 사용자 기능 추가
- `fix`: 버그 수정
- `refactor`: 동작 변경 없는 구조 개선
- `test`: 테스트 추가/수정
- `docs`: 문서 변경
- `chore`: 빌드, 설정, 의존성, 보조 작업

## scope 작성 기준

현재 디렉토리 구조 기준으로 아래처럼 맞추는 것이 자연스럽습니다.

- `worker`
- `flow`
- `domain`
- `cmd`
- `docs`

예시:

```text
feat(flow): create project with ULID
test(flow): use require assertions
docs(docs): add commit convention research
```

## 제목 작성 규칙

- 가능하면 50자 안팎으로 짧게 작성
- 명령형 현재 시제 사용
- 첫 글자는 불필요하게 대문자로 시작하지 않음
- 마침표를 붙이지 않음

## 본문 작성 규칙

- 왜 바꿨는지와 무엇이 달라졌는지 설명
- 줄 길이는 72자 안팎으로 감쌈
- breaking change나 이슈 번호는 footer로 분리 가능

## 참고 자료

- Conventional Commits 공식 소개: <https://www.conventionalcommits.org/en/about>
- Git `git-commit` 문서: <https://git-scm.com/docs/git-commit>
- Git `SubmittingPatches` 문서: <https://git-scm.com/docs/SubmittingPatches>
