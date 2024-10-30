package server

import "regexp"
// проверки регулярных выражений и уникальности логина
const (
	usernameR = `^[a-zA-Z0-9_]{3,20}$`
	passR = `^[A-Za-z\d]{8,}$`
) 
/*
passR:
^ — указывает на начало строки.
(?=.*[A-Za-z]) — утверждение, что строка должна содержать хотя бы одну букву:
[A-Za-z] — любая буква (как заглавная, так и строчная).
.* — позволяет находить буквы в любом месте строки.
(?=.*\d) — утверждение, что строка должна содержать хотя бы одну цифру:
\d — любая цифра.
[A-Za-z\d]{8,} — сама строка, состоящая из букв и цифр, и имеет длину не менее 8 символов.
$ — указывает на конец строки.
*/

func isValidUsername(username string)bool {
	match, _ := regexp.MatchString(usernameR, username)
	return match
}

func isValidPass(password string)bool {
	if match, _ := regexp.MatchString(passR, password); !match {
		return false
	}
	if match, _ := regexp.MatchString(`[A-Za-z]`, password); !match {
		return false
	}
	if match, _ := regexp.MatchString(`\d`, password); !match {
		return false
    }
	return true
}
