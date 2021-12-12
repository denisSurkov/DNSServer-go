## Как запустить:
1) [Скачать go](https://golang.org/dl/)
2) Перейти в корень проекта
3)
```
go mod download
go build 
sudo ./DNSServer
```
Для выполнения спросит root (запускается на порту 53, как и все dns сервера)

## При выполнении использовались 
- https://datatracker.ietf.org/doc/html/rfc1035
- Так много ссылок, что забыл ввести подсчет

Для проверки работы использовал:
``dig @localhost +retry=1 google.com``
