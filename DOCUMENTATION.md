# kvault User Guide

**kvault** is a personal knowledge management system. It keeps your notes, web pages, and PDF documents in one place, tags them automatically, and lets you instantly find what you need with full-text search.

Everything you add to kvault becomes searchable: note text, the content of saved web pages, and text from uploaded PDF files.

> [!WARNING]
> **About privacy and security — read this before you start.**
>
> kvault is a system for **personal self-hosting**. This means:
>
> - **The server administrator sees everything.** The owner of the server running kvault has full access to your data: note content, uploaded files, tags, and even your API key. Data is stored **without encryption**. Only entrust your data to a server operated by you or by someone you fully trust.
> - **API keys are per-device and expire.** Every login issues its own key; a key stops working if unused for longer than the inactivity period (30 days by default). If a key ends up in someone else's hands (unlikely for a personal, non-exposed server, but still), revoke access in the account settings: delete the specific key or end sessions on other devices.
> - **Do not expose the service to the open internet.** kvault has no built-in protection against attacks. Keep the server in a closed perimeter: use a VPN that admits only your trusted devices, or restrict access with a firewall. See [SECURITY.md](./SECURITY.md).

---

## Contents

- [Quick start](#quick-start)
- [Registration and login](#registration-and-login)
- [Notes](#notes)
  - [Text notes](#text-notes)
  - [URL notes](#url-notes)
- [Files](#files)
- [Full-text search (FTS)](#full-text-search-fts)
- [Filters and sorting](#filters-and-sorting)
- [Tags](#tags)
- [Stopwords and auto-tagging](#stopwords-and-auto-tagging)
- [Trash bin](#trash-bin)
- [Language and theme](#language-and-theme)
- [API reference (Swagger)](#api-reference-swagger)

---

## Quick start

A typical kvault workflow:

1. **Register** and log in.
2. **Add knowledge** — create a text note, save a web page by URL, or upload a PDF file.
3. kvault **automatically indexes** the content and **assigns tags**.
4. **Find what you need** via search, filter by tags.
5. When needed, **manage tags and stopwords** to make auto-tagging more precise.

---

## Registration and login

1. On the login page, click the registration link.
2. Enter a **username** (at least 3 characters) and a **password**.
3. After registration you are logged in automatically.

> **How authorization works.** Each login issues a separate **API key** for that device. The web UI stores it in the browser and attaches it automatically — you never enter the key yourself. You only need the key if you want to call the API directly (see [Swagger](#api-reference-swagger)). A key expires if unused for longer than the inactivity period (30 days by default); active keys are extended automatically.

In the account settings you can:

- **change your password**;
- **manage active keys** — view the device list, delete a specific key, end the current session, or log out on all other devices;
- **delete your account** along with all its data.

---

## Notes

Notes come in two types: **text** and **URL**. Both are created via the same create button — pick the tab you need in the dialog that opens.

### Text notes

1. Click the note creation button.
2. On the **"Text"** tab, enter a **title** and press Enter (or the create button).
3. An editor opens — notes are written in **Markdown** with a preview.

Right after saving, the note is indexed for search and tags are automatically derived from its content.

### URL notes

1. In the creation dialog, switch to the **"Link"** tab.
2. Enter a **title** and the **page address**.
3. kvault fetches the page itself and extracts:
   - the **title, description, and site name** (from the page metadata);
   - a **preview image** — the site's preview is shown right in the note;
   - the **main page text** — it becomes searchable.

The extracted text is indexed and participates in search and auto-tagging just like regular notes. A saved article is therefore findable by words from its content, not just its title.

---

## Files

The **"Files"** section is for documents. **PDF files** are supported.

**How to upload:**

- click the upload button and pick a file, **or**
- drag a PDF straight into the file list area.

After upload the file is processed in the background: text is **extracted** from the PDF and becomes searchable. While processing is in progress, the file card shows the corresponding status.

**The label on the file card** shows whether the text is ready for search:

| Label | Meaning |
| --- | --- |
| **Searchable** | Text was extracted from the file and indexed — the file is findable by its content. |
| **Not searchable** | Text could not be extracted (e.g. the PDF consists of scanned images with no text layer). The file is stored and downloadable, but not findable by content. |

A file can be **opened for viewing**, **downloaded**, or **deleted** (to the trash bin).

---

## Full-text search (FTS)

The search field is available in the "Notes" and "Files" sections. Search covers **all content**: titles, note text, extracted web page text, and text from PDFs. Results update as you type.

The query syntax is close to a familiar Google-like search:

### Multiple words narrow the result (logical AND)

The more words you type, the narrower the result: entries are found that contain **all** of the given words (in any order, not necessarily adjacent).

```
project documentation
```
→ finds entries containing **both** "project" **and** "documentation".

### Prefix search

You don't have to type a word out in full — kvault matches by the **beginning** of a word. Handy for different word forms and search-as-you-type.

```
docum
```
→ finds "**docum**ent", "**docum**entation", "**docum**entary", etc.

```
prog lang
```
→ finds entries containing a word starting with "prog…" (program, programming) **and** a word starting with "lang…".

### Punctuation is ignored

Periods, commas, brackets, and other symbols inside a query are discarded — a word is reduced to letters, digits, and hyphens. The queries `e-mail`, `e-mail,` and `e mail` give practically the same result.

### Case doesn't matter

`Postgres`, `postgres`, and `POSTGRES` are the same thing.

> **What search does not do.** It's not Google in the full sense: special operators (quotes for exact phrases, `OR`, minus to exclude words) are **not supported**. Multiple words are always combined with AND. If nothing is found — drop extra words or shorten them to a prefix.

**Usage examples:**

| What you're looking for | Query |
| --- | --- |
| Notes about configuring nginx | `nginx config` |
| The article that mentioned Kubernetes | `kuber` |
| The document about the annual tax report | `tax report` |

Search can be combined with [tag filters](#filters-and-sorting) — e.g. find all entries tagged "work" that mention the word "deadline".

---

## Filters and sorting

Below the search bar are the result controls.

### Tag filter

The **"Tags"** button opens your tag list with its own search. Check one or more tags:

- with **multiple** tags selected, entries are shown that have **at least one** of the selected tags;
- the counter on the button shows how many tags are currently selected;
- the **"Clear"** button resets the selection.

### Sorting

The dropdown sets the sort field, and the arrow button toggles the direction (ascending ↑ / descending ↓).

**Notes** can be sorted by:
- modification date,
- creation date,
- title.

**Files** can be sorted by:
- creation date,
- file name,
- size.

> With an active search, results are additionally ranked by relevance — the best matches come first.

All conditions work together: search + tag filter + sorting apply simultaneously.

---

## Tags

Tags are labels for grouping and filtering entries. They are either **automatic** (assigned by kvault, see [auto-tagging](#stopwords-and-auto-tagging)) or **manual**.

### Tags on the note page

On an open note's page you can:

- **add a tag** — the plus button opens the list; pick an existing tag or **create a new one** by typing its name;
- **remove a tag** — clicking an attached tag removes it from the note;
- **assign tags automatically** — see below.

### Managing all tags (Settings → Tags)

The **"Tags"** tab in settings shows the full list of your tags:

- **search** and **sorting** (by name, creation date, modification date);
- next to each tag, the number of **entries using it** is shown;
- **renaming** — click the tag's name;
- **deletion** — removes the tag and detaches it from all entries;
- **manual creation** of a new tag.

---

## Stopwords and auto-tagging

These are two related features: auto-tagging picks tags from text, and stopwords control which words must **not** become tags.

### How auto-tagging works

When you create a note (or request tag suggestions manually), kvault analyzes its text and picks several of the most characteristic words as tags. The algorithm:

1. Takes all of the entry's indexed text (title + content + extracted link text).
2. Counts **how often** each word occurs.
3. **Filters out** of the candidates:
   - words that are too short (3 letters or fewer);
   - words containing more than just letters (numbers, codes, part numbers);
   - **stopwords** (see below).
4. The remaining words are sorted by frequency, and the **most frequent** become tags.

> **Enough text required.** Very short entries are not tagged automatically — a couple of words is not enough to determine the topic. This also applies to URL notes: if too little meaningful text could be extracted from the page, no auto-tags are created.

### Manual tag suggestions

The note page has an **auto-tagging** action with a configurable **number of tags** (5 by default). Set the desired count and run it — kvault adds suitable tags to the note. Handy when you've extended a note and want to refresh its tag set. New tags immediately appear in the global tag list.

### Stopwords (Settings → Stopwords)

Stopwords are words **excluded from auto-tagging**. Without them, function words like "this", "which", "the", "and" would become tags.

kvault ships with a ready-made list of common stopwords for **Russian and English**. You can adjust the list to your needs:

- **Enable / disable** a word with the toggle. A disabled stopword can become a tag again. This also works for the built-in list — if you want a "work" tag, for example, disable the corresponding stopword.
- **Add your own** stopword — type it and confirm. Useful for words that are frequent in your domain but useless as tags.
- **Delete** — only **your own** words can be deleted; built-in ones can't be deleted, but can be disabled.
- The **source label** shows where a word came from:

| Source | Meaning |
| --- | --- |
| **Built-in** | From kvault's bundled list. Can be disabled, can't be deleted. |
| **Custom** | Added by you. Can be disabled and deleted. |

The list supports search, sorting (by word, source, modification date), and filtering by source.

> **When changes take effect.** Stopwords are applied at tag-selection time. Already assigned tags are not recalculated automatically when the stopword list changes — to refresh a specific note's tags, run manual tag suggestions for it.

---

## Trash bin

Deleted notes and files go to the **trash bin** instead of being erased immediately. From the bin an entry can be:

- **restored** to its previous place;
- **deleted permanently** (one entry at a time, or by emptying the whole bin).

This protects deletion from accidents.

---

## Language and theme

The language and theme switches are at the top of the UI.

**UI language.** The button with the language code (`EN` / `RU` / `JA`) opens the selection:

- 🇬🇧 English
- 🇷🇺 Русский
- 🇯🇵 日本語

**Theme.** The sun/moon icon button toggles the **light** and **dark** themes.

Both choices are saved in the browser and applied automatically on your next visit.

---

## API reference (Swagger)

Besides the web UI, kvault provides a **REST API**. If you want to integrate kvault with your scripts or applications, the full interactive documentation is available at:

```
http://<your-server-address>/swagger
```

It lists all available operations (notes, files, tags, stopwords, authorization) with parameter descriptions and the ability to send a test request. API calls use your personal **API key** in the `Authorization` header.
