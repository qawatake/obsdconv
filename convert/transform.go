package convert

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/scan"
)

func currentLine(raw []rune, ptr int) (linenum int) {
	return strings.Count(string(raw[:ptr]), "\n") + 1
}

func TransformNone(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
	return 1, raw[ptr : ptr+1], nil
}

func TransformInternalLinkFunc(t InternalLinkTransformer) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scan.ScanInternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		link, err := t.TransformInternalLink(content)
		if err != nil {
			return 0, nil, errors.Wrap(err, "genExternalLink failed in TransformInternalLinkFunc")
		}
		return advance, []rune(link), nil
	}
}

func defaultTransformInternalLinkFunc(db PathDB) TransformerFunc {
	return TransformInternalLinkFunc(newInternalLinkTransformerImpl(db))
}

func TransformEmnbedsFunc(t EmbedsTransformer) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scan.ScanEmbeds(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		link, err := t.TransformEmbeds(content)
		if err != nil {
			return 0, nil, errors.Wrap(err, "genExternalLink failed in TransformEmbedsFunc")
		}
		return advance, []rune(link), nil
	}
}

func defaultTransformEmbedsFunc(db PathDB) TransformerFunc {
	return TransformEmnbedsFunc(newEmbedsTransformerImpl(db))
}

func TransformExternalLinkFunc(t ExternalLinkTransformer) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, displayName, ref, title := scan.ScanExternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}

		externalLink, err := t.TransformExternalLink(displayName, ref, title)
		if err != nil {
			return 0, nil, errors.Wrap(err, "t.TransformExternalLink failed")
		}
		return advance, []rune(externalLink), nil
	}
}

func defaultTransformExternalLinkFunc(db PathDB) TransformerFunc {
	return TransformExternalLinkFunc(newExternalLinkTransformerImpl(db))
}

func TransformInternalLinkToPlain(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
	advance, content := scan.ScanInternalLink(raw, ptr)
	if advance == 0 {
		return 0, nil, nil
	}
	if content == "" { // [[ ]] はスキップ
		return advance, nil, nil
	}

	identifier, displayName := splitDisplayName(content)
	if displayName != "" {
		return advance, []rune(displayName), nil
	}
	fileId, fragments, err := splitFragments(identifier)
	if err != nil {
		return 0, nil, errors.Wrap(err, "splitFragments failed in TransformInternalLinkFunc")
	}

	linktext := buildLinkText(displayName, fileId, fragments)
	return advance, []rune(linktext), nil
}

func TransformExternalLinkToPlain(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
	advance, displayName, _, _ := scan.ScanExternalLink(raw, ptr)
	if advance == 0 {
		return 0, nil, nil
	}
	return advance, []rune(displayName), nil
}

type InternalLinkTransformer interface {
	TransformInternalLink(content string) (externalLink string, err error)
}

type InternalLinkTransformerImpl struct {
	PathDB
}

func newInternalLinkTransformerImpl(db PathDB) *InternalLinkTransformerImpl {
	return &InternalLinkTransformerImpl{
		PathDB: db,
	}
}

func (t *InternalLinkTransformerImpl) TransformInternalLink(content string) (externalLink string, err error) {
	if content == "" {
		return "", nil // [[ ]] はスキップ
	}

	identifier, displayName := splitDisplayName(content)
	fileId, fragments, err := splitFragments(identifier)
	if err != nil {
		return "", errors.Wrap(err, "splitFragments failed")
	}
	path, err := t.Get(fileId)
	if err != nil {
		return "", errors.Wrap(err, "PathDB.Get failed")
	}

	linktext := buildLinkText(displayName, fileId, fragments)
	var ref string
	if fragments == nil {
		ref = path
	} else {
		ref = path + "#" + formatAnchor(fragments[len(fragments)-1])
	}

	return fmt.Sprintf("[%s](%s)", linktext, ref), nil
}

type EmbedsTransformer interface {
	TransformEmbeds(content string) (embeddedLink string, err error)
}

type EmbedsTransformerImpl struct {
	PathDB
}

func newEmbedsTransformerImpl(db PathDB) *EmbedsTransformerImpl {
	return &EmbedsTransformerImpl{
		PathDB: db,
	}
}

func (t *EmbedsTransformerImpl) TransformEmbeds(content string) (emnbeddedLink string, err error) {
	if content == "" {
		return "", nil // [[ ]] はスキップ
	}

	identifier, displayName := splitDisplayName(content)
	fileId, fragments, err := splitFragments(identifier)
	if err != nil {
		return "", errors.Wrap(err, "splitFragments failed")
	}
	path, err := t.Get(fileId)
	if err != nil {
		return "", errors.Wrap(err, "PathDB.Get failed")
	}

	linktext := buildLinkText(displayName, fileId, fragments)
	var ref string
	if fragments == nil {
		ref = path
	} else {
		ref = path + "#" + formatAnchor(fragments[len(fragments)-1])
	}

	return fmt.Sprintf("![%s](%s)", linktext, ref), nil
}

type ExternalLinkTransformer interface {
	TransformExternalLink(displayName, ref string, title string) (externalLink string, err error)
}

type ExternalLinkTransformerImpl struct {
	PathDB
}

func newExternalLinkTransformerImpl(db PathDB) *ExternalLinkTransformerImpl {
	return &ExternalLinkTransformerImpl{
		PathDB: db,
	}
}

func (t *ExternalLinkTransformerImpl) TransformExternalLink(displayName, ref string, title string) (externalLink string, err error) {
	u, err := url.Parse(ref)
	if err != nil {
		return "", newErrTransformf(ERR_KIND_UNEXPECTED, "url.Parse failed: %v", err)
	}

	// ref = 通常のリンク
	if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		if title == "" {
			return fmt.Sprintf("[%s](%s)", displayName, ref), nil
		} else {
			return fmt.Sprintf("[%s](%s \"%s\")", displayName, ref, title), nil
		}
	}

	// ref = obsidian URI (obsidian://open?...)
	// ignore vault query (?vault=...) and path query (?path=...)
	// resolve path by using file query (?file=...) and PathDB
	if u.Scheme == "obsidian" && u.Host == "open" {
		q := u.Query()
		fileId := q.Get("file")
		if fileId == "" {
			return "", newErrTransformf(ERR_KIND_NO_REF_SPECIFIED_IN_OBSIDIAN_URL, "no ref file specified in obsidian url: %s", ref)
		}
		path, err := t.Get(fileId)
		if err != nil {
			return "", errors.Wrap(err, "PathDB.Get failed")
		}
		if title == "" {
			return fmt.Sprintf("[%s](%s)", displayName, path), nil
		} else {
			return fmt.Sprintf("[%s](%s \"%s\")", displayName, path, title), nil
		}
	}

	// ref = obsidian URI (obsidian://vault/my_vault/my_note)
	// ignore vault parameter
	// resolve path by using file parameter and PathDB
	if u.Scheme == "obsidian" && u.Host == "vault" {
		segments := strings.Split(u.Path, "/")
		if len(segments) != 3 {
			return "", newErrTransformf(ERR_KIND_INVALID_SHORTHAND_OBSIDIAN_URL, "invalid shorthand obsidian url: %s", ref)
		}
		fileId := segments[2]
		path, err := t.Get(fileId)
		if err != nil {
			return "", errors.Wrap(err, "PathDB.Get failed")
		}
		if title == "" {
			return fmt.Sprintf("[%s](%s)", displayName, path), nil
		} else {
			return fmt.Sprintf("[%s](%s \"%s\")", displayName, path, title), nil
		}
	}

	// ref = fileId
	if u.Scheme == "" && u.Host == "" {
		fileId, fragments, err := splitFragments(ref)
		if err != nil {
			return "", errors.Wrap(err, "splitFragments failed")
		}
		path, err := t.Get(fileId)
		if err != nil {
			return "", errors.Wrap(err, "PathDB.Get failed")
		}
		var newref string
		if fragments == nil {
			newref = path
		} else {
			newref = path + "#" + strings.Join(fragments, "#")
		}
		if title == "" {
			return fmt.Sprintf("[%s](%s)", displayName, newref), nil
		} else {
			return fmt.Sprintf("[%s](%s \"%s\")", displayName, newref, title), nil
		}
	}

	return "", newErrTransformf(ERR_KIND_UNEXPECTED_HREF, "unexpected href: %s", ref)
}

func formatAnchor(rawAnchor string) (anchor string) {
	loweredAnchor := strings.ToLower(rawAnchor)
	rawRunes := []rune(loweredAnchor)
	runes := make([]rune, 0, len(rawRunes))
	for _, r := range rawRunes {
		if r == ' ' {
			runes = append(runes, '-')
			continue
		}
		if emojis.in(r) {
			continue
		}
		if isSymbolToBeIgnored(r) {
			continue
		}
		runes = append(runes, r)
	}
	return string(runes)
}

func isSymbolToBeIgnored(r rune) bool {
	for _, symbol := range []rune("!@#$%^&*()+|~=\\`[]{};':\",./<>?") {
		if r == symbol {
			return true
		}
	}
	for _, zenkaku := range []rune("！＠＃＄％＾＆＊（）＋｜〜＝￥｀「」｛｝；’：”、。・＜＞？　【】『』《》〔〕［］‹›«»〘〙〚〛") {
		if r == zenkaku {
			return true
		}
	}
	return false
}

// https://go.dev/play/p/PRRmhnXNHf
type emoji []struct {
	lo rune
	hi rune
}

func (ee emoji) in(r rune) bool {
	for _, e := range ee {
		if e.lo <= r && r <= e.hi {
			return true
		}
	}
	return false
}

var emojis = emoji{
	{0x0023, 0x0023},   //  (#️)       number sign
	{0x002A, 0x002A},   //  (*️)       asterisk
	{0x0030, 0x0039},   //  (0️..9️)    digit zero..digit nine
	{0x00A9, 0x00A9},   //  (©️)       copyright
	{0x00AE, 0x00AE},   //  (®️)       registered
	{0x203C, 0x203C},   //  (‼️)       double exclamation mark
	{0x2049, 0x2049},   //  (⁉️)       exclamation question mark
	{0x2122, 0x2122},   //  (™️)       trade mark
	{0x2139, 0x2139},   //  (ℹ️)       information
	{0x2194, 0x2199},   //  (↔️..↙️)    left-right arrow..down-left arrow
	{0x21A9, 0x21AA},   //  (↩️..↪️)    right arrow curving left..left arrow curving right
	{0x231A, 0x231B},   //  (⌚..⌛)    watch..hourglass done
	{0x2328, 0x2328},   //  (⌨️)       keyboard
	{0x23CF, 0x23CF},   //  (⏏️)       eject button
	{0x23E9, 0x23F3},   //  (⏩..⏳)    fast-forward button..hourglass not done
	{0x23F8, 0x23FA},   //  (⏸️..⏺️)    pause button..record button
	{0x24C2, 0x24C2},   //  (Ⓜ️)       circled M
	{0x25AA, 0x25AB},   //  (▪️..▫️)    black small square..white small square
	{0x25B6, 0x25B6},   //  (▶️)       play button
	{0x25C0, 0x25C0},   //  (◀️)       reverse button
	{0x25FB, 0x25FE},   //  (◻️..◾)    white medium square..black medium-small square
	{0x2600, 0x2604},   //  (☀️..☄️)    sun..comet
	{0x260E, 0x260E},   //  (☎️)       telephone
	{0x2611, 0x2611},   //  (☑️)       ballot box with check
	{0x2614, 0x2615},   //  (☔..☕)    umbrella with rain drops..hot beverage
	{0x2618, 0x2618},   //  (☘️)       shamrock
	{0x261D, 0x261D},   //  (☝️)       index pointing up
	{0x2620, 0x2620},   //  (☠️)       skull and crossbones
	{0x2622, 0x2623},   //  (☢️..☣️)    radioactive..biohazard
	{0x2626, 0x2626},   //  (☦️)       orthodox cross
	{0x262A, 0x262A},   //  (☪️)       star and crescent
	{0x262E, 0x262F},   //  (☮️..☯️)    peace symbol..yin yang
	{0x2638, 0x263A},   //  (☸️..☺️)    wheel of dharma..smiling face
	{0x2640, 0x2640},   //  (♀️)       female sign
	{0x2642, 0x2642},   //  (♂️)       male sign
	{0x2648, 0x2653},   //  (♈..♓)    Aries..Pisces
	{0x2660, 0x2660},   //  (♠️)       spade suit
	{0x2663, 0x2663},   //  (♣️)       club suit
	{0x2665, 0x2666},   //  (♥️..♦️)    heart suit..diamond suit
	{0x2668, 0x2668},   //  (♨️)       hot springs
	{0x267B, 0x267B},   //  (♻️)       recycling symbol
	{0x267F, 0x267F},   //  (♿)       wheelchair symbol
	{0x2692, 0x2697},   //  (⚒️..⚗️)    hammer and pick..alembic
	{0x2699, 0x2699},   //  (⚙️)       gear
	{0x269B, 0x269C},   //  (⚛️..⚜️)    atom symbol..fleur-de-lis
	{0x26A0, 0x26A1},   //  (⚠️..⚡)    warning..high voltage
	{0x26AA, 0x26AB},   //  (⚪..⚫)    white circle..black circle
	{0x26B0, 0x26B1},   //  (⚰️..⚱️)    coffin..funeral urn
	{0x26BD, 0x26BE},   //  (⚽..⚾)    soccer ball..baseball
	{0x26C4, 0x26C5},   //  (⛄..⛅)    snowman without snow..sun behind cloud
	{0x26C8, 0x26C8},   //  (⛈️)       cloud with lightning and rain
	{0x26CE, 0x26CE},   //  (⛎)       Ophiuchus
	{0x26CF, 0x26CF},   //  (⛏️)       pick
	{0x26D1, 0x26D1},   //  (⛑️)       rescue worker’s helmet
	{0x26D3, 0x26D4},   //  (⛓️..⛔)    chains..no entry
	{0x26E9, 0x26EA},   //  (⛩️..⛪)    shinto shrine..church
	{0x26F0, 0x26F5},   //  (⛰️..⛵)    mountain..sailboat
	{0x26F7, 0x26FA},   //  (⛷️..⛺)    skier..tent
	{0x26FD, 0x26FD},   //  (⛽)       fuel pump
	{0x2702, 0x2702},   //  (✂️)       scissors
	{0x2705, 0x2705},   //  (✅)       white heavy check mark
	{0x2708, 0x2709},   //  (✈️..✉️)    airplane..envelope
	{0x270A, 0x270B},   //  (✊..✋)    raised fist..raised hand
	{0x270C, 0x270D},   //  (✌️..✍️)    victory hand..writing hand
	{0x270F, 0x270F},   //  (✏️)       pencil
	{0x2712, 0x2712},   //  (✒️)       black nib
	{0x2714, 0x2714},   //  (✔️)       heavy check mark
	{0x2716, 0x2716},   //  (✖️)       heavy multiplication x
	{0x271D, 0x271D},   //  (✝️)       latin cross
	{0x2721, 0x2721},   //  (✡️)       star of David
	{0x2728, 0x2728},   //  (✨)       sparkles
	{0x2733, 0x2734},   //  (✳️..✴️)    eight-spoked asterisk..eight-pointed star
	{0x2744, 0x2744},   //  (❄️)       snowflake
	{0x2747, 0x2747},   //  (❇️)       sparkle
	{0x274C, 0x274C},   //  (❌)       cross mark
	{0x274E, 0x274E},   //  (❎)       cross mark button
	{0x2753, 0x2755},   //  (❓..❕)    question mark..white exclamation mark
	{0x2757, 0x2757},   //  (❗)       exclamation mark
	{0x2763, 0x2764},   //  (❣️..❤️)    heavy heart exclamation..red heart
	{0x2795, 0x2797},   //  (➕..➗)    heavy plus sign..heavy division sign
	{0x27A1, 0x27A1},   //  (➡️)       right arrow
	{0x27B0, 0x27B0},   //  (➰)       curly loop
	{0x27BF, 0x27BF},   //  (➿)       double curly loop
	{0x2934, 0x2935},   //  (⤴️..⤵️)    right arrow curving up..right arrow curving down
	{0x2B05, 0x2B07},   //  (⬅️..⬇️)    left arrow..down arrow
	{0x2B1B, 0x2B1C},   //  (⬛..⬜)    black large square..white large square
	{0x2B50, 0x2B50},   //  (⭐)       white medium star
	{0x2B55, 0x2B55},   //  (⭕)       heavy large circle
	{0x3030, 0x3030},   //  (〰️)       wavy dash
	{0x303D, 0x303D},   //  (〽️)       part alternation mark
	{0x3297, 0x3297},   //  (㊗️)       Japanese “congratulations” button
	{0x3299, 0x3299},   //  (㊙️)       Japanese “secret” button
	{0x1F004, 0x1F004}, //  (🀄)       mahjong red dragon
	{0x1F0CF, 0x1F0CF}, //  (🃏)       joker
	{0x1F170, 0x1F171}, //  (🅰️..🅱️)    A button (blood type)..B button (blood type)
	{0x1F17E, 0x1F17E}, //  (🅾️)       O button (blood type)
	{0x1F17F, 0x1F17F}, //  (🅿️)       P button
	{0x1F18E, 0x1F18E}, //  (🆎)       AB button (blood type)
	{0x1F191, 0x1F19A}, //  (🆑..🆚)    CL button..VS button
	{0x1F1E6, 0x1F1FF}, //  (🇦..🇿)    regional indicator symbol letter a..regional indicator symbol letter z
	{0x1F201, 0x1F202}, //  (🈁..🈂️)    Japanese “here” button..Japanese “service charge” button
	{0x1F21A, 0x1F21A}, //  (🈚)       Japanese “free of charge” button
	{0x1F22F, 0x1F22F}, //  (🈯)       Japanese “reserved” button
	{0x1F232, 0x1F23A}, //  (🈲..🈺)    Japanese “prohibited” button..Japanese “open for business” button
	{0x1F250, 0x1F251}, //  (🉐..🉑)    Japanese “bargain” button..Japanese “acceptable” button
	{0x1F300, 0x1F320}, //  (🌀..🌠)    cyclone..shooting star
	{0x1F321, 0x1F321}, //  (🌡️)       thermometer
	{0x1F324, 0x1F32C}, //  (🌤️..🌬️)    sun behind small cloud..wind face
	{0x1F32D, 0x1F32F}, //  (🌭..🌯)    hot dog..burrito
	{0x1F330, 0x1F335}, //  (🌰..🌵)    chestnut..cactus
	{0x1F336, 0x1F336}, //  (🌶️)       hot pepper
	{0x1F337, 0x1F37C}, //  (🌷..🍼)    tulip..baby bottle
	{0x1F37D, 0x1F37D}, //  (🍽️)       fork and knife with plate
	{0x1F37E, 0x1F37F}, //  (🍾..🍿)    bottle with popping cork..popcorn
	{0x1F380, 0x1F393}, //  (🎀..🎓)    ribbon..graduation cap
	{0x1F396, 0x1F397}, //  (🎖️..🎗️)    military medal..reminder ribbon
	{0x1F399, 0x1F39B}, //  (🎙️..🎛️)    studio microphone..control knobs
	{0x1F39E, 0x1F39F}, //  (🎞️..🎟️)    film frames..admission tickets
	{0x1F3A0, 0x1F3C4}, //  (🎠..🏄)    carousel horse..person surfing
	{0x1F3C5, 0x1F3C5}, //  (🏅)       sports medal
	{0x1F3C6, 0x1F3CA}, //  (🏆..🏊)    trophy..person swimming
	{0x1F3CB, 0x1F3CE}, //  (🏋️..🏎️)    person lifting weights..racing car
	{0x1F3CF, 0x1F3D3}, //  (🏏..🏓)    cricket game..ping pong
	{0x1F3D4, 0x1F3DF}, //  (🏔️..🏟️)    snow-capped mountain..stadium
	{0x1F3E0, 0x1F3F0}, //  (🏠..🏰)    house..castle
	{0x1F3F3, 0x1F3F5}, //  (🏳️..🏵️)    white flag..rosette
	{0x1F3F7, 0x1F3F7}, //  (🏷️)       label
	{0x1F3F8, 0x1F3FF}, //  (🏸..🏿)    badminton..dark skin tone
	{0x1F400, 0x1F43E}, //  (🐀..🐾)    rat..paw prints
	{0x1F43F, 0x1F43F}, //  (🐿️)       chipmunk
	{0x1F440, 0x1F440}, //  (👀)       eyes
	{0x1F441, 0x1F441}, //  (👁️)       eye
	{0x1F442, 0x1F4F7}, //  (👂..📷)    ear..camera
	{0x1F4F8, 0x1F4F8}, //  (📸)       camera with flash
	{0x1F4F9, 0x1F4FC}, //  (📹..📼)    video camera..videocassette
	{0x1F4FD, 0x1F4FD}, //  (📽️)       film projector
	{0x1F4FF, 0x1F4FF}, //  (📿)       prayer beads
	{0x1F500, 0x1F53D}, //  (🔀..🔽)    shuffle tracks button..down button
	{0x1F549, 0x1F54A}, //  (🕉️..🕊️)    om..dove
	{0x1F54B, 0x1F54E}, //  (🕋..🕎)    kaaba..menorah
	{0x1F550, 0x1F567}, //  (🕐..🕧)    one o’clock..twelve-thirty
	{0x1F56F, 0x1F570}, //  (🕯️..🕰️)    candle..mantelpiece clock
	{0x1F573, 0x1F579}, //  (🕳️..🕹️)    hole..joystick
	{0x1F57A, 0x1F57A}, //  (🕺)       man dancing
	{0x1F587, 0x1F587}, //  (🖇️)       linked paperclips
	{0x1F58A, 0x1F58D}, //  (🖊️..🖍️)    pen..crayon
	{0x1F590, 0x1F590}, //  (🖐️)       hand with fingers splayed
	{0x1F595, 0x1F596}, //  (🖕..🖖)    middle finger..vulcan salute
	{0x1F5A4, 0x1F5A4}, //  (🖤)       black heart
	{0x1F5A5, 0x1F5A5}, //  (🖥️)       desktop computer
	{0x1F5A8, 0x1F5A8}, //  (🖨️)       printer
	{0x1F5B1, 0x1F5B2}, //  (🖱️..🖲️)    computer mouse..trackball
	{0x1F5BC, 0x1F5BC}, //  (🖼️)       framed picture
	{0x1F5C2, 0x1F5C4}, //  (🗂️..🗄️)    card index dividers..file cabinet
	{0x1F5D1, 0x1F5D3}, //  (🗑️..🗓️)    wastebasket..spiral calendar
	{0x1F5DC, 0x1F5DE}, //  (🗜️..🗞️)    clamp..rolled-up newspaper
	{0x1F5E1, 0x1F5E1}, //  (🗡️)       dagger
	{0x1F5E3, 0x1F5E3}, //  (🗣️)       speaking head
	{0x1F5E8, 0x1F5E8}, //  (🗨️)       left speech bubble
	{0x1F5EF, 0x1F5EF}, //  (🗯️)       right anger bubble
	{0x1F5F3, 0x1F5F3}, //  (🗳️)       ballot box with ballot
	{0x1F5FA, 0x1F5FA}, //  (🗺️)       world map
	{0x1F5FB, 0x1F5FF}, //  (🗻..🗿)    mount fuji..moai
	{0x1F600, 0x1F600}, //  (😀)       grinning face
	{0x1F601, 0x1F610}, //  (😁..😐)    beaming face with smiling eyes..neutral face
	{0x1F611, 0x1F611}, //  (😑)       expressionless face
	{0x1F612, 0x1F614}, //  (😒..😔)    unamused face..pensive face
	{0x1F615, 0x1F615}, //  (😕)       confused face
	{0x1F616, 0x1F616}, //  (😖)       confounded face
	{0x1F617, 0x1F617}, //  (😗)       kissing face
	{0x1F618, 0x1F618}, //  (😘)       face blowing a kiss
	{0x1F619, 0x1F619}, //  (😙)       kissing face with smiling eyes
	{0x1F61A, 0x1F61A}, //  (😚)       kissing face with closed eyes
	{0x1F61B, 0x1F61B}, //  (😛)       face with tongue
	{0x1F61C, 0x1F61E}, //  (😜..😞)    winking face with tongue..disappointed face
	{0x1F61F, 0x1F61F}, //  (😟)       worried face
	{0x1F620, 0x1F625}, //  (😠..😥)    angry face..sad but relieved face
	{0x1F626, 0x1F627}, //  (😦..😧)    frowning face with open mouth..anguished face
	{0x1F628, 0x1F62B}, //  (😨..😫)    fearful face..tired face
	{0x1F62C, 0x1F62C}, //  (😬)       grimacing face
	{0x1F62D, 0x1F62D}, //  (😭)       loudly crying face
	{0x1F62E, 0x1F62F}, //  (😮..😯)    face with open mouth..hushed face
	{0x1F630, 0x1F633}, //  (😰..😳)    anxious face with sweat..flushed face
	{0x1F634, 0x1F634}, //  (😴)       sleeping face
	{0x1F635, 0x1F640}, //  (😵..🙀)    dizzy face..weary cat face
	{0x1F641, 0x1F642}, //  (🙁..🙂)    slightly frowning face..slightly smiling face
	{0x1F643, 0x1F644}, //  (🙃..🙄)    upside-down face..face with rolling eyes
	{0x1F645, 0x1F64F}, //  (🙅..🙏)    person gesturing NO..folded hands
	{0x1F680, 0x1F6C5}, //  (🚀..🛅)    rocket..left luggage
	{0x1F6CB, 0x1F6CF}, //  (🛋️..🛏️)    couch and lamp..bed
	{0x1F6D0, 0x1F6D0}, //  (🛐)       place of worship
	{0x1F6D1, 0x1F6D2}, //  (🛑..🛒)    stop sign..shopping cart
	{0x1F6E0, 0x1F6E5}, //  (🛠️..🛥️)    hammer and wrench..motor boat
	{0x1F6E9, 0x1F6E9}, //  (🛩️)       small airplane
	{0x1F6EB, 0x1F6EC}, //  (🛫..🛬)    airplane departure..airplane arrival
	{0x1F6F0, 0x1F6F0}, //  (🛰️)       satellite
	{0x1F6F3, 0x1F6F3}, //  (🛳️)       passenger ship
	{0x1F6F4, 0x1F6F6}, //  (🛴..🛶)    kick scooter..canoe
	{0x1F6F7, 0x1F6F8}, //  (🛷..🛸)    sled..flying saucer
	{0x1F910, 0x1F918}, //  (🤐..🤘)    zipper-mouth face..sign of the horns
	{0x1F919, 0x1F91E}, //  (🤙..🤞)    call me hand..crossed fingers
	{0x1F91F, 0x1F91F}, //  (🤟)       love-you gesture
	{0x1F920, 0x1F927}, //  (🤠..🤧)    cowboy hat face..sneezing face
	{0x1F928, 0x1F92F}, //  (🤨..🤯)    face with raised eyebrow..exploding head
	{0x1F930, 0x1F930}, //  (🤰)       pregnant woman
	{0x1F931, 0x1F932}, //  (🤱..🤲)    breast-feeding..palms up together
	{0x1F933, 0x1F93A}, //  (🤳..🤺)    selfie..person fencing
	{0x1F93C, 0x1F93E}, //  (🤼..🤾)    people wrestling..person playing handball
	{0x1F940, 0x1F945}, //  (🥀..🥅)    wilted flower..goal net
	{0x1F947, 0x1F94B}, //  (🥇..🥋)    1st place medal..martial arts uniform
	{0x1F94C, 0x1F94C}, //  (🥌)       curling stone
	{0x1F950, 0x1F95E}, //  (🥐..🥞)    croissant..pancakes
	{0x1F95F, 0x1F96B}, //  (🥟..🥫)    dumpling..canned food
	{0x1F980, 0x1F984}, //  (🦀..🦄)    crab..unicorn face
	{0x1F985, 0x1F991}, //  (🦅..🦑)    eagle..squid
	{0x1F992, 0x1F997}, //  (🦒..🦗)    giraffe..cricket
	{0x1F9C0, 0x1F9C0}, //  (🧀)       cheese wedge
	{0x1F9D0, 0x1F9E6}, //  (🧐..🧦)    face with monocle..socks
}
