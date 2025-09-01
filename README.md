````markdown
# 🖼️ ASCII Art Generator на Go

Консольная утилита для преобразования изображений и GIF-анимаций в цветной ASCII-арт.
Поддерживает **PNG/JPG/GIF**, вывод в **терминал** и в **Markdown** (для GitHub профиля).

---

## 🚀 Установка

```bash
git clone https://github.com/yourname/ascii-art-go.git
cd ascii-art-go
go build -o ascii-art .
```
````

---

## 📸 Использование

Поддерживаются форматы **PNG**, **JPG** и **GIF**.

### 1. Статичное изображение в терминале (цветом)

```bash
./ascii-art -file image.png -width 120
```

### 2. Анимированный GIF — бесконечная прокрутка в терминале

```bash
./ascii-art -file cat.gif -width 90 -loop=true
```

С фиксированным FPS:

```bash
./ascii-art -file cat.gif -width 90 -fps 12
```

### 3. Генерация ASCII для GitHub (Markdown)

```bash
./ascii-art -file image.png -width 100 -markdown -out README_snippet.md
```

В `README_snippet.md` появится блок:

````markdown
```
  .:-=+*#%@
   :-=+*#%@
    =+*#%@
     *#%@
```
````

### 4. Черно-белый ASCII (# и пробел)

```bash
./ascii-art -file image.png -width 100 -bw
```

### 5. Ключи

| Флаг        | Описание                                       |
| ----------- | ---------------------------------------------- |
| `-file`     | путь к файлу (png/jpg/gif)                     |
| `-width`    | ширина ASCII (по умолчанию 100)                |
| `-color`    | цветной вывод ANSI (true/false)                |
| `-bw`       | черно-белый вывод (# и пробел)                 |
| `-markdown` | режим Markdown (удаляет ANSI цвета)            |
| `-out`      | сохранить результат в файл                     |
| `-fps`      | FPS для GIF (0 = использовать задержки из GIF) |
| `-loop`     | крутить GIF бесконечно (true/false)            |

---

## 📖 Пример для GitHub профиля

Вставь ASCII-арт в свой `README.md` так:

---

## ⚡ Возможности

- ✅ Цветной ASCII в терминале
- ✅ Поддержка анимированных GIF
- ✅ Экспорт в Markdown для README
- ✅ Настройка ширины и FPS
- ✅ Черно-белый режим (# и пробел)
