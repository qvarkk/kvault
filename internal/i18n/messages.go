package i18n

var messages = map[string]map[string]string{
	"err.bad_request": {
		LocaleEN: "One or more fields have validation errors. Please check and try again.",
		LocaleRU: "Одно или несколько полей содержат ошибки валидации. Пожалуйста, проверьте данные.",
		LocaleJA: "1つ以上のフィールドにバリデーションエラーがあります。確認してください。",
	},
	"err.unauthorized": {
		LocaleEN: "Wrong credentials. Please check and try again.",
		LocaleRU: "Неверные учётные данные. Пожалуйста, проверьте и попробуйте снова.",
		LocaleJA: "認証情報が正しくありません。確認して再試行してください。",
	},
	"err.forbidden": {
		LocaleEN: "Access to the requested entity is forbidden.",
		LocaleRU: "Доступ к запрашиваемому ресурсу запрещён.",
		LocaleJA: "要求されたリソースへのアクセスは禁止されています。",
	},
	"err.not_found": {
		LocaleEN: "The requested resource was not found.",
		LocaleRU: "Запрашиваемый ресурс не найден.",
		LocaleJA: "要求されたリソースが見つかりませんでした。",
	},
	"err.unprocessable": {
		LocaleEN: "The request could not be processed. Please check your input.",
		LocaleRU: "Запрос не может быть обработан. Пожалуйста, проверьте входные данные.",
		LocaleJA: "リクエストを処理できませんでした。入力内容を確認してください。",
	},
	"err.internal": {
		LocaleEN: "An internal server error occurred.",
		LocaleRU: "Произошла внутренняя ошибка сервера.",
		LocaleJA: "内部サーバーエラーが発生しました。",
	},
	"err.invalid_api_key": {
		LocaleEN: "Invalid or missing API key.",
		LocaleRU: "Неверный или отсутствующий API-ключ.",
		LocaleJA: "APIキーが無効または不足しています。",
	},
	"err.access_forbidden": {
		LocaleEN: "Access forbidden.",
		LocaleRU: "Доступ запрещён.",
		LocaleJA: "アクセスが禁止されています。",
	},
	"err.invalid_credentials": {
		LocaleEN: "Invalid credentials provided.",
		LocaleRU: "Предоставлены неверные учётные данные.",
		LocaleJA: "認証情報が正しくありません。",
	},
	"err.user_not_found": {
		LocaleEN: "User not found.",
		LocaleRU: "Пользователь не найден.",
		LocaleJA: "ユーザーが見つかりませんでした。",
	},
	"err.user_already_exists": {
		LocaleEN: "User with this username already exists.",
		LocaleRU: "Пользователь с таким именем пользователя уже существует.",
		LocaleJA: "このユーザー名はすでに使用されています。",
	},
	"err.item_not_found": {
		LocaleEN: "Item with given ID does not exist.",
		LocaleRU: "Элемент с указанным ID не существует.",
		LocaleJA: "指定されたIDのアイテムは存在しません。",
	},
	"err.file_not_found": {
		LocaleEN: "File with given ID does not exist.",
		LocaleRU: "Файл с указанным ID не существует.",
		LocaleJA: "指定されたIDのファイルは存在しません。",
	},
	"err.stopword_not_found": {
		LocaleEN: "This stopword does not exist.",
		LocaleRU: "Стоп-слово не найдено.",
		LocaleJA: "このストップワードは存在しません。",
	},
	"err.stopword_already_exists": {
		LocaleEN: "This stopword already exists.",
		LocaleRU: "Это стоп-слово уже существует.",
		LocaleJA: "このストップワードはすでに存在します。",
	},
	"err.tag_not_found": {
		LocaleEN: "This tag does not exist.",
		LocaleRU: "Тег не найден.",
		LocaleJA: "このタグは存在しません。",
	},
	"err.tag_already_exists": {
		LocaleEN: "This tag already exists.",
		LocaleRU: "Тег уже существует.",
		LocaleJA: "このタグはすでに存在します。",
	},
	"err.pdf_format": {
		LocaleEN: "File should be of a PDF content type.",
		LocaleRU: "Файл должен быть в формате PDF.",
		LocaleJA: "ファイルはPDF形式でなければなりません。",
	},
	"err.insufficient_content": {
		LocaleEN: "Not enough content to generate the requested number of tags.",
		LocaleRU: "Недостаточно контента для генерации указанного количества тегов.",
		LocaleJA: "要求されたタグ数を生成するコンテンツが不足しています。",
	},
}

func Translate(key, locale string) string {
	locs, ok := messages[key]
	if !ok {
		return ""
	}
	if msg, ok := locs[locale]; ok && msg != "" {
		return msg
	}
	return locs[LocaleEN]
}
