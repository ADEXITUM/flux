Flux - фреймворк, предоставляющий Gin-like интерфейс для взаимодействия. 
Разрабатывался именно для того, чтобы подменить его в проектах где используется Gin, при этом получить дополнительный функционал

Кастомное поле Engine.authFunc позволяет модифицировать Context необходимыми данными. Сама структура контекста легко меняется.

Добавлено:
1. Защита роутов аутентификацией или RBAC
2. time.Time в Context для удобного использования в метриках
3. ParseBody в тело контекста при входящем запросе. Учитывает особенности Multipart/FormData