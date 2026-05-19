package i18n

import (
	"fmt"
	"strings"
)

var validationFormats = map[string]map[string]string{
	"required": {
		LocaleEN: "%s is required",
		LocaleRU: "Поле %s обязательно для заполнения",
		LocaleJA: "%sは必須です",
	},
	"min": {
		LocaleEN: "%s must be at least %s characters",
		LocaleRU: "%s должно содержать не менее %s символов",
		LocaleJA: "%sは%s文字以上でなければなりません",
	},
	"max": {
		LocaleEN: "%s cannot be longer than %s characters",
		LocaleRU: "%s не может быть длиннее %s символов",
		LocaleJA: "%sは%s文字以下でなければなりません",
	},
	"email": {
		LocaleEN: "Invalid email format",
		LocaleRU: "Неверный формат email",
		LocaleJA: "メールアドレスの形式が正しくありません",
	},
	"uuid4": {
		LocaleEN: "%s must follow uuid4 format",
		LocaleRU: "%s должно соответствовать формату uuid4",
		LocaleJA: "%sはuuid4形式でなければなりません",
	},
	"oneof": {
		LocaleEN: "%s must be one of: %s",
		LocaleRU: "%s должно быть одним из: %s",
		LocaleJA: "%sは次のいずれかでなければなりません: %s",
	},
	"_default": {
		LocaleEN: "%s is not valid",
		LocaleRU: "%s недействительно",
		LocaleJA: "%sは無効です",
	},
}

func FormatValidation(tag, field, param, locale string) string {
	formats, ok := validationFormats[tag]
	if !ok {
		formats = validationFormats["_default"]
	}
	tpl, ok := formats[locale]
	if !ok || tpl == "" {
		tpl = formats[LocaleEN]
	}
	switch strings.Count(tpl, "%s") {
	case 0:
		return tpl
	case 1:
		return fmt.Sprintf(tpl, field)
	default:
		return fmt.Sprintf(tpl, field, param)
	}
}
