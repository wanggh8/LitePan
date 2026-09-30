package rules

import (
	"fmt"
	"regexp"
	"strings"
)

func SanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "",
		"\\", "",
		":", "：",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
	)
	return strings.TrimSpace(replacer.Replace(name))
}

func IsSameGeneratedName(currentName, generatedName string) bool {
	return strings.TrimSpace(currentName) == strings.TrimSpace(generatedName)
}

func BuildFolderName(parsed ParsedMedia, tmdbID string) string {
	title := strings.TrimSpace(parsed.Title)
	if title == "" {
		return ""
	}
	tag := ""
	if tmdbID != "" {
		tag = fmt.Sprintf("{tmdb-%s}", tmdbID)
	}
	parts := []string{title}
	if parsed.Year != nil {
		parts = append(parts, fmt.Sprintf("(%d)", *parsed.Year))
	}
	if tag != "" {
		parts = append(parts, tag)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

// BuildTargetFilename 生成目标文件名（不含扩展名）。
// 电影格式：标题 (年份) [标记]
// 剧集格式：标题.年份.SxxExx（点分隔，年份无括号）
func BuildTargetFilename(parsed ParsedMedia, marker, tmdbID string) string {
	title := strings.TrimSpace(parsed.Title)
	if title == "" {
		return ""
	}

	season := asFirstInt(parsed.Season)
	episode := asFirstInt(parsed.Episode)
	isTV := season != nil && episode != nil

	tag := ""
	if !IsMarkerOff(marker) {
		switch strings.TrimSpace(marker) {
		case "tmdb", "tmdbid":
			if tmdbID != "" {
				tag = fmt.Sprintf("{tmdb-%s}", tmdbID)
			}
		default:
			if m := strings.TrimSpace(marker); m != "" {
				tag = fmt.Sprintf("[%s]", m)
			}
		}
	}

	if isTV {
		// 剧集：标题.年份.SxxExx（点分隔，年份无括号，标记紧跟年份后）
		parts := []string{title}
		if parsed.Year != nil {
			parts = append(parts, fmt.Sprintf("%d", *parsed.Year))
		}
		if tag != "" {
			parts = append(parts, tag)
		}
		digits := 2
		if parsed.EpisodeDigits >= 3 || *episode >= 100 {
			digits = 3
		}
		return strings.Join(parts, ".") + fmt.Sprintf(".S%02dE%0*d", *season, digits, *episode)
	}

	// 电影：标题 (年份) [标记]（空格分隔，年份带括号）
	parts := []string{title}
	if parsed.Year != nil {
		parts = append(parts, fmt.Sprintf("(%d)", *parsed.Year))
	}
	if tag != "" {
		parts = append(parts, tag)
	}
	return strings.Join(parts, " ")
}

func BuildDisplayTitle(tmdbTitle, tmdbOriginal, fallbackTitle string) string {
	if tmdbTitle != "" {
		return tmdbTitle
	}
	return fallbackTitle
}

func FitFilenameBytes(filename, tmdbLang string) string {
	if len([]byte(filename)) <= MaxFilenameBytes {
		return filename
	}
	// 逐步删减 " - tag1.tag2..." 部分的末尾 tag
	sep := " - "
	sepIdx := strings.LastIndex(filename, sep)
	if sepIdx < 0 {
		return filename
	}
	afterSep := filename[sepIdx+len(sep):]
	dotIdx := strings.LastIndex(afterSep, ".")
	if dotIdx < 0 {
		return filename
	}
	ext := afterSep[dotIdx:]  // ".mkv"
	tags := afterSep[:dotIdx] // "2160p.HDR10.TrueHD"
	tagParts := strings.Split(tags, ".")
	for i := len(tagParts); i > 0; i-- {
		newTags := strings.Join(tagParts[:i], ".")
		short := filename[:sepIdx] + sep + newTags + ext
		if len([]byte(short)) <= MaxFilenameBytes {
			return short
		}
	}
	return filename[:sepIdx] + ext
}

func IsAlreadyOrganized(filename, marker string) bool {
	if IsMarkerOff(marker) {
		return looksLikeOrganizedStructure(filename) || FindTMDBIDInName(filename) != ""
	}
	m := strings.TrimSpace(marker)
	switch m {
	case "tmdb", "tmdbid":
		return FindTMDBIDInName(filename) != ""
	default:
		return strings.Contains(filename, fmt.Sprintf("[%s]", m))
	}
}

func looksLikeOrganizedStructure(filename string) bool {
	return organizedStructureRe.MatchString(strings.TrimSpace(filename))
}

func IsMarkerOff(marker string) bool {
	switch strings.ToLower(strings.TrimSpace(marker)) {
	case "", "0", "off", "none", "no", "false":
		return true
	default:
		return false
	}
}

var (
	bracketTagRe = regexp.MustCompile(`\[([^\]]+)\]`)
)
