# gitverse-notifier

Сервис для обработки вебхуков gitverse, создан из говна и палок для личных нужд и от безделья.

## Запуск

### Локально

```bash
make init_configs
# нужно указать переменные в config.yml JIRA_TOKEN, GITVERSE_BASE_URL и тд
make build
./build/gitverse_notifier -f config.yml
```

### Docker

```bash
make docker
```

## Конфигурация

Конифиги можно укзать в config.yml или через переменные окружения.
Базовые конфиги:

| Ключ | Описание | Значение по умолчанию |
|---|---|---|
| `BIND_HOST` | Хост | `0.0.0.0` |
| `HTTPD_PORT` | Порт | `9001` |
| `LOG_LEVEL` | Уровень логирования | `INFO` |
| `LOG_FILE_NAME` | Имя файла с логами | `gitverse-notifier.log` |
| `JIRA_URL` | URL для Jira | — |
| `JIRA_TOKEN` | Токен для Jira | — |
| `JIRA_USERNAME` / `JIRA_PASSWORD` | Логин и пароль для джиры - альтернатива токену | — |
| `GITVERSE_BASE_URL` | Ссылка на гитверс для использования в уведомлениях | — |


## Конфигурация репозиториев

Файлы с настройками репозиториев лежат в `configs/repositories/*.yml`, для примера есть `configs/repositories/default.yml`.

```yaml
repository: any
# Базовый URL gitverse для ссылок
url: ""
# Коды проектов Jira
allowed_jira_projects: []
```

## Конфигурация правил уведомления

Файлы с правилами уведомлений лежат в `configs/actions/*.yml`, для примера есть `configs/actions/default.yml`. В них указывается имя репозитория (any для всех) и список правил.
Например:

```yaml
repository: any

actions:
  - on: pull_request.opened # Код события
    action: jira.comment_issue # Код действия
    template: jira/pr_opened # Название шаблона

  - on: pull_request.opened
    action: telegram.notify
    template: telegram/pr_opened

  - on: branch.push
    branch: main
    action: telegram.notify
    template: telegram/branch_push

  - on: pull_request.comment
    action: telegram.notify
    skip_empty: true # gitverse бывает присылает события с пустым сообщением, опция для их пропуска
    template: telegram/pr_comment
```

### Список событий

Внутри приложения есть ряд типов событий, которые могут использоваться в action->on.
События из gitverse мапятся в них на основании значений заголовком X-Gitverse-Event и X-Gitverse-Event-Type (и иногда поля action).

| Событие | Описание |
|---|---|
| `pull_request.opened` / `.edited` / `.closed` / `synchronized` | Жизненый цикл PR |
| `pull_request.review_requested` | Запрос ревью PR |
| `pull_request.review_approved` | PR одобрен |
| `pull_request.review_rejected` | PR отклонен |
| `pull_request.review_comment` | Комментарий для PR (но сам комментарий не передается) |
| `pull_request.comment` | Комментарий для PR (с текстом) |
| `branch.push` | В ветку запушены изменения |
| `branch.created` / `branch.deleted` | Ветка создана / удалена |
| `cicd.status` | Изменение состояния CICD |


### Список действий

| Действие | Описание |
|---|---|
| `jira.comment_issue` | Создает комментарий для задачи в Jira |
| `telegram.notify` | Пока заглушка |


## Шаблоны

Файлы внутри `configs/templates/**/*.tmpl` используются как go template.
Регистрируются с именами вида:
- `configs/templates/jira/pr_opened.tmpl` -> `jira/pr_opened`
- `configs/templates/telegram/pr_comment.tmpl` -> `telegram/pr_comment`

Имена шаблонов можно указываться в action->template.

Пример шаблона

```
{{jiraLink .Sender.Name .Sender.URL}} открыл запрос на слияние{{if .PullRequest.Number}} #{{.PullRequest.Number}}{{end}}: {{jiraLink .PullRequest.Title .PullRequest.URL}} ({{jiraLink .Repository .RepositoryURL}})
{{- if .Branch}}
Ветка: {{jiraLink .Branch .BranchURL}}
{{- end}}
```

### Как парсятся номера задач Jira

- Из залоговка PR - берутся похожие на номера строки из начала, например  `JIRA-1 Fix` или `JIRA-1 JIRA-2 Fix`
- Из имени ветки (branch или ref)
- Ключи фильтруются по `allowed_jira_projects` из настроек репозитория

## В планах

- Отправка уведомлений в тг
- Расширить логирование
- Добавить фильтрацию по результату для cicd.status
- добавить обратку skip_empty
- добавить интеграцию с gitverse (обновление заголовка PR для указания ссылки на задачу)
