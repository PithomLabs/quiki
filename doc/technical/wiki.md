# wiki
--
    import "github.com/cooper/quiki/wiki"


## Usage

```go
const (
	// CategoryTypeImage is a type of category that tracks which pages use an image.
	CategoryTypeImage CategoryType = "image"

	// CategoryTypeModel is a type of category that tracks which pages use a model.
	CategoryTypeModel = "model"

	// CategoryTypePage is a type of category that tracks which pages reference another page.
	CategoryTypePage = "page"
)
```

#### type Category

```go
type Category struct {

	// category path
	Path string `json:"-"`

	// category filename, including the .cat extension
	File string `json:"-"`

	// category name without extension
	Name string `json:"name,omitempty"`

	// human-readable category title
	Title string `json:"title,omitempty"`

	// time when the category was created
	Created     *time.Time `json:"created,omitempty"`
	CreatedHTTP string     `json:"created_http,omitempty"` // HTTP formatted

	// time when the category was last modified.
	// this is updated when pages are added and deleted
	Modified     *time.Time `json:"modified,omitempty"`
	ModifiedHTTP string     `json:"modified_http,omitempty"` // HTTP formatted

	// pages in the category. keys are filenames
	Pages map[string]CategoryEntry `json:"pages,omitempty"`

	// when true, the category is preserved even when no pages remain
	Preserve bool `json:"preserve,omitempty"`

	// if applicable, this is the type of special category.
	// for normal categories, this is empty
	Type CategoryType `json:"type,omitempty"`

	// for CategoryTypePage, this is the info for the tracked page
	PageInfo *wikifier.PageInfo `json:"page_info,omitempty"`

	// for CategoryTypeImage, this is the info for the tracked image
	ImageInfo *struct {
		Width  int `json:"width,omitempty"`
		Height int `json:"height,omitempty"`
	} `json:"image_info,omitempty"`
}
```

A Category is a collection of pages pertaining to a topic.

A page can belong to many categories. Category memberships and metadta are
stored in JSON manifests.

#### func (*Category) AddPage

```go
func (cat *Category) AddPage(w *Wiki, page *wikifier.Page)
```
AddPage adds a page to a category.

If the page already belongs and any information has changed, the category is
updated. If force is true,

#### func (*Category) Exists

```go
func (cat *Category) Exists() bool
```
Exists returns whether a category currently exists.

#### type CategoryEntry

```go
type CategoryEntry struct {

	// time at which the page metadata in this category file was last updated.
	// this is compared against page file modification time
	Asof *time.Time `json:"asof,omitempty"`

	// embedded page info
	// note this info is accurate only as of the Asof time
	wikifier.PageInfo

	// for CategoryTypeImage, an array of image dimensions used on this page.
	// dimensions are guaranteed to be positive integers. the number of elements will
	// always be even, since each occurrence of the image produces two (width and then height)
	Dimensions [][]int `json:"dimensions,omitempty"`

	// for CategoryTypePage, an array of line numbers on which the tracked page is
	// referenced on the page described by this entry
	Lines []int `json:"lines,omitempty"`
}
```

A CategoryEntry describes a page that belongs to a category.

#### type CategoryType

```go
type CategoryType string
```

CategoryType describes the type of a Category.

#### type DisplayCategoryPosts

```go
type DisplayCategoryPosts struct {

	// DisplayPage results
	// overrides the Category Pages field
	Pages []DisplayPage `json:"pages,omitempty"`

	// the page number (first page = 0)
	PageN int `json:"page_n"`

	// the total number of pages
	NumPages int `json:"num_pages"`

	// this is the combined CSS for all pages we're displaying
	CSS string `json:"css,omitempty"`

	// all other fields are inherited from the category itself
	*Category
}
```

DisplayCategoryPosts represents a category result to display.

#### type DisplayError

```go
type DisplayError struct {
	// a human-readable error string. sensitive info is never
	// included, so this may be shown to users
	Error string

	// a more detailed human-readable error string that MAY contain
	// sensitive data. can be used for debugging and logging but should
	// not be presented to users
	DetailedError string

	// HTTP status code. if zero, 404 should be used
	Status int

	// true if the error occurred during parsing
	ParseError bool

	// true if the content cannot be displayed because it has
	// not yet been published for public access
	Draft bool
}
```

DisplayError represents an error result to display.

#### type DisplayImage

```go
type DisplayImage struct {

	// basename of the scaled image file
	File string `json:"file,omitempty"`

	// absolute path to the scaled image.
	// this file should be served to the user
	Path string `json:"path,omitempty"`

	// absolute path to the full-size image.
	// if the full-size image is being displayed, same as Path
	FullsizePath string `json:"fullsize_path,omitempty"`

	// image type
	// 'png' or 'jpeg'
	ImageType string `json:"image_type,omitempty"`

	// mime 'image/png' or 'image/jpeg'
	// suitable for the Content-Type header
	Mime string `json:"mime,omitempty"`

	// bytelength of image data
	// suitable for use in the Content-Length header
	Length int64 `json:"length,omitempty"`

	// time when the image was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the image file itself.
	Modified     *time.Time `json:"modified,omitempty"`
	ModifiedHTTP string     `json:"modified_http,omitempty"` // HTTP format for Last-Modified

	// true if the content being sered was read from a cache file.
	// opposite of Generated
	FromCache bool `json:"cached,omitempty"`

	// true if the content being served was just generated.
	// opposite of FromCache
	Generated bool `json:"generated,omitempty"`

	// true if the content generated in order to fulfill this request was
	// written to cache. this can only been true when Generated is true
	CacheGenerated bool `json:"cache_gen,omitempty"`
}
```

DisplayImage represents an image to display.

#### type DisplayPage

```go
type DisplayPage struct {

	// basename of the page, with the extension
	File string `json:"file,omitempty"`

	// basename of the page, without the extension
	Name string `json:"name,omitempty"`

	// absolute file path of the page
	Path string `json:"path,omitempty"`

	// the page content (HTML)
	Content wikifier.HTML `json:"content,omitempty"`

	// time when the page was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the page file itself.
	Modified     *time.Time `json:"modified,omitempty"`
	ModifiedHTTP string     `json:"modified_http,omitempty"` // HTTP formatted for Last-Modified

	// CSS generated for the page from style{} blocks
	CSS string `json:"css,omitempty"`

	// true if this content was read from a cache file. opposite of Generated
	FromCache bool `json:"cached,omitempty"`

	// true if the content being served was just generated on the fly.
	// opposite of FromCache
	Generated bool `json:"generated,omitempty"`

	// true if this request resulted in the writing of a new cache file.
	// this can only be true if Generated is true
	CacheGenerated bool `json:"cache_gen,omitempty"`

	// true if this request resulted in the writing of a text file.
	// this can only be true if Generated is true
	TextGenerated bool `json:"text_gen,omitempty"`

	// true if the page has not yet been published for public viewing.
	// this only occurs when it is specified that serving drafts is OK,
	// since normally a draft page instead results in a DisplayError.
	Draft bool `json:"draft,omitempty"`

	// warnings produced by the parser
	Warnings []string `json:"warnings,omitempty"`

	// time when the page was created, as extracted from
	// the special @page.created variable
	Created     *time.Time `json:"created,omitempty"`
	CreatedHTTP string     `json:"created_http,omitempty"` // HTTP formatted

	// name of the page author, as extracted from the special @page.author
	// variable
	Author string `json:"author,omitempty"`

	// list of categories the page belongs to, without the '.cat' extension
	Categories []string `json:"categories,omitempty"`

	// page title as extracted from the special @page.title variable, including
	// any possible HTML-encoded formatting
	FmtTitle wikifier.HTML `json:"fmt_title,omitempty"`

	// like FmtTitle except that all text formatting has been stripped.
	// suitable for use in the <title> tag
	Title string `json:"title,omitempty"`
}
```

DisplayPage represents a page result to display.

#### type DisplayRedirect

```go
type DisplayRedirect struct {

	// a relative or absolute URL to which the request should redirect,
	// suitable for use in a Location header
	Redirect string
}
```

DisplayRedirect represents a page redirect to follow.

#### type ImageInfo

```go
type ImageInfo struct {
	File       string     `json:"file"`               // filename
	Width      int        `json:"width,omitempty"`    // full-size width
	Height     int        `json:"height,omitempty"`   // full-size height
	Created    *time.Time `json:"created,omitempty"`  // creation time
	Modified   *time.Time `json:"modified,omitempty"` // modify time
	Dimensions [][]int    `json:"-"`                  // dimensions used throughout the wiki
}
```

ImageInfo represents a full-size image on the wiki.

#### type SizedImage

```go
type SizedImage struct {
	// for example 100x200-myimage@3x.png
	Width, Height int    // 100, 200 (dimensions as requested)
	Scale         int    // 3 (scale as requested)
	Name          string // myimage (name without extension)
	Ext           string // png (extension)
}
```

SizedImage represents an image in specific dimensions.

#### func  SizedImageFromName

```go
func SizedImageFromName(name string) SizedImage
```
SizedImageFromName returns a SizedImage given an image name.

#### func (SizedImage) FullName

```go
func (img SizedImage) FullName() string
```
FullName returns the image name with true dimensions.

#### func (SizedImage) FullNameNE

```go
func (img SizedImage) FullNameNE() string
```
FullNameNE is like FullName but without the extension.

#### func (SizedImage) FullSizeName

```go
func (img SizedImage) FullSizeName() string
```
FullSizeName returns the name of the full-size image.

#### func (SizedImage) ScaleName

```go
func (img SizedImage) ScaleName() string
```
ScaleName returns the image name with dimensions and scale.

#### func (SizedImage) TrueHeight

```go
func (img SizedImage) TrueHeight() int
```
TrueHeight returns the actual image height when the Scale is taken into
consideration.

#### func (SizedImage) TrueWidth

```go
func (img SizedImage) TrueWidth() int
```
TrueWidth returns the actual image width when the Scale is taken into
consideration.

#### type Wiki

```go
type Wiki struct {
	ConfigFile        string
	PrivateConfigFile string
	Opt               wikifier.PageOpt
}
```

A Wiki represents a quiki website.

#### func  NewWiki

```go
func NewWiki(conf, privateConf string) (*Wiki, error)
```
NewWiki creates a Wiki given the public and private configuration files.

#### func (*Wiki) DisplayCategoryPosts

```go
func (w *Wiki) DisplayCategoryPosts(catName string, pageN int) interface{}
```
DisplayCategoryPosts returns the display result for a category.

#### func (*Wiki) DisplayImage

```go
func (w *Wiki) DisplayImage(name string) interface{}
```
DisplayImage returns the display result for an image.

#### func (*Wiki) DisplayPage

```go
func (w *Wiki) DisplayPage(name string) interface{}
```
DisplayPage returns the display result for a page.

#### func (*Wiki) DisplayPageDraft

```go
func (w *Wiki) DisplayPageDraft(name string, draftOK bool) interface{}
```
DisplayPageDraft returns the display result for a page.

Unlike DisplayPage, if draftOK is true, the content is served even if it is
marked as draft.

#### func (*Wiki) DisplaySizedImage

```go
func (w *Wiki) DisplaySizedImage(img SizedImage) interface{}
```
DisplaySizedImage returns the display result for an image in specific
dimensions.

#### func (*Wiki) DisplaySizedImageGenerate

```go
func (w *Wiki) DisplaySizedImageGenerate(img SizedImage, generateOK bool) interface{}
```
DisplaySizedImageGenerate returns the display result for an image in specific
dimensions and allows images to be generated in any dimension.

#### func (*Wiki) GetCategory

```go
func (w *Wiki) GetCategory(name string) *Category
```
GetCategory loads or creates a category.

#### func (*Wiki) GetSpecialCategory

```go
func (w *Wiki) GetSpecialCategory(name string, typ CategoryType) *Category
```
GetSpecialCategory loads or creates a special category given the type.

#### func (*Wiki) ImageInfo

```go
func (w *Wiki) ImageInfo(name string) (info ImageInfo)
```
ImageInfo returns info for an image given its full-size name.

#### func (*Wiki) Images

```go
func (w *Wiki) Images() map[string]ImageInfo
```
Images returns info about all the images in the wiki.

#### func (*Wiki) NewPage

```go
func (w *Wiki) NewPage(name string) *wikifier.Page
```
NewPage creates a Page given its name and configures it for use with this Wiki.

#### func (*Wiki) Pregenerate

```go
func (w *Wiki) Pregenerate()
```
Pregenerate simulates requests for all wiki resources such that content caches
can be pregenerated and stored.