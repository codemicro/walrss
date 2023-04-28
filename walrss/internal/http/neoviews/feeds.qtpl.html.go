// Code generated by qtc from "feeds.qtpl.html". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

package neoviews

import "github.com/codemicro/walrss/walrss/internal/db"

import "github.com/codemicro/walrss/walrss/internal/urls"

import "github.com/codemicro/walrss/walrss/internal/http/neoviews/internal/components"

import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

type FeedsPageArgs struct {
	DigestsEnabled bool
	SelectedDay    db.SendDay
	SelectedTime   int
	Feeds          []*db.Feed
	Categories     []*db.Category
}

func StreamFeedsPage(qw422016 *qt422016.Writer, args *FeedsPageArgs) {
	qw422016.N().S(`
<!DOCTYPE html>
<html lang="en">
`)
	components.StreamHead(qw422016, "Feeds")
	qw422016.N().S(`
<body>
`)
	components.StreamNavigation(qw422016, menuItems, menuItemFeeds)
	qw422016.N().S(`

`)
	components.StreamBeginMain(qw422016)
	qw422016.N().S(`

<div class="container">
    <h1 class="title"><i class="bi bi-rss-fill"></i> My Feeds</h1>

    <div class="card">
        `)
	if !args.DigestsEnabled {
		qw422016.N().S(`
            <p style="margin-top: 0;"><span class="warning-text">Warning:</span> you have digests disabled. No emails will be sent until you re-enable digests in your settings.</p>
        `)
	}
	qw422016.N().S(`
        `)
	StreamRenderFeedTabsAndTable(qw422016, args.Feeds, args.Categories, "", false)
	qw422016.N().S(`
    </div>
</div>

`)
	components.StreamEndMain(qw422016)
	qw422016.N().S(`

<div id="modal-target"></div>
`)
	components.StreamToast(qw422016)
	qw422016.N().S(`

</body>
</html>
`)
}

func WriteFeedsPage(qq422016 qtio422016.Writer, args *FeedsPageArgs) {
	qw422016 := qt422016.AcquireWriter(qq422016)
	StreamFeedsPage(qw422016, args)
	qt422016.ReleaseWriter(qw422016)
}

func FeedsPage(args *FeedsPageArgs) string {
	qb422016 := qt422016.AcquireByteBuffer()
	WriteFeedsPage(qb422016, args)
	qs422016 := string(qb422016.B)
	qt422016.ReleaseByteBuffer(qb422016)
	return qs422016
}

func StreamRenderFeedTabsAndTable(qw422016 *qt422016.Writer, feeds []*db.Feed, categories []*db.Category, activeCategoryID string, oob bool) {
	qw422016.N().S(`
    <div id="feeds" `)
	if oob {
		qw422016.N().S(`hx-swap-oob="outerHTML"`)
	}
	qw422016.N().S(`>
        <div class="tabs">
            <div class="tab `)
	if activeCategoryID == "" {
		qw422016.N().S(`active`)
	}
	qw422016.N().S(`" hx-get="`)
	qw422016.E().S(urls.FeedCategoryTab)
	qw422016.N().S(`" hx-target="#feeds" hx-swap="outerHTML">(no category)</div>
            `)
	for _, category := range categories {
		qw422016.N().S(`
                <div class="tab `)
		if category.ID == activeCategoryID {
			qw422016.N().S(`active`)
		}
		qw422016.N().S(`" hx-get="`)
		qw422016.E().S(urls.FeedCategoryTab)
		qw422016.N().S(`?category=`)
		qw422016.E().S(category.ID)
		qw422016.N().S(`" hx-target="#feeds" hx-swap="outerHTML">`)
		qw422016.E().S(category.Name)
		qw422016.N().S(`</div>
            `)
	}
	qw422016.N().S(`
            <div class="tab" hx-get="`)
	qw422016.E().S(urls.NewCategory)
	qw422016.N().S(`" hx-target="#modal-target">+</div>
            <div class="filler-line"></div>
        </div>
        <div class="tab-box">
            <div class="flex-horizontal">
                <button class="button green"
                        hx-get="`)
	qw422016.E().S(urls.NewFeedItem)
	qw422016.N().S(`?category=`)
	qw422016.E().S(activeCategoryID)
	qw422016.N().S(`"
                        hx-target="#modal-target"
                >Add new feed</button>
                `)
	if activeCategoryID != "" {
		qw422016.N().S(`
                    <button class="button">Edit category</button>
                    <button class="button">Delete category</button>
                `)
	}
	qw422016.N().S(`
            </div>

            <hr>

            `)
	if len(feeds) == 0 {
		qw422016.N().S(`
            <p><i>There's nothing here!</i></p>
            `)
	}
	qw422016.N().S(`

            <table class="table">
                `)
	for _, feed := range feeds {
		qw422016.N().S(`
                <tr>
                    <td>`)
		qw422016.E().S(feed.Name)
		qw422016.N().S(`</td>
                    <td>`)
		qw422016.E().S(feed.URL)
		qw422016.N().S(`</td>
                    <td>
                        <div class="flex-horizontal float-right">
                            <button class="button inline"
                                hx-get="`)
		qw422016.E().S(urls.Expand(urls.EditFeedItem, feed.ID))
		if activeCategoryID != "" {
			qw422016.N().S(`?category=`)
			qw422016.E().S(activeCategoryID)
		}
		qw422016.N().S(`"
                                hx-target="#modal-target"
                            >Edit</button>
                            <button class="button inline"
                                hx-delete="`)
		qw422016.E().S(urls.Expand(urls.EditFeedItem, feed.ID))
		if activeCategoryID != "" {
			qw422016.N().S(`?category=`)
			qw422016.E().S(activeCategoryID)
		}
		qw422016.N().S(`"
                                hx-confirm="This will permanently delete this item. Are you sure?"
                            >Delete</button>
                        </div>
                    </td>
                </tr>
                `)
	}
	qw422016.N().S(`
            </table>
        </div>
    </div>
`)
}

func WriteRenderFeedTabsAndTable(qq422016 qtio422016.Writer, feeds []*db.Feed, categories []*db.Category, activeCategoryID string, oob bool) {
	qw422016 := qt422016.AcquireWriter(qq422016)
	StreamRenderFeedTabsAndTable(qw422016, feeds, categories, activeCategoryID, oob)
	qt422016.ReleaseWriter(qw422016)
}

func RenderFeedTabsAndTable(feeds []*db.Feed, categories []*db.Category, activeCategoryID string, oob bool) string {
	qb422016 := qt422016.AcquireByteBuffer()
	WriteRenderFeedTabsAndTable(qb422016, feeds, categories, activeCategoryID, oob)
	qs422016 := string(qb422016.B)
	qt422016.ReleaseByteBuffer(qb422016)
	return qs422016
}

type FragmentNewFeedArgs struct {
	CurrentCategoryID string
	Categories        []*db.Category
}

func StreamFragmentNewFeed(qw422016 *qt422016.Writer, args *FragmentNewFeedArgs) {
	qw422016.N().S(`
    `)
	components.StreamBeginModal(qw422016)
	qw422016.N().S(`
    <h2><i class="bi bi-pencil-square"></i> Add New Feed</h2>
    <p class="warning-text">// TODO: category</p>
    <form>
        <div class="form-grid">
            <label for="new-feed-name">Feed name</label>
            <input type="text" id="new-feed-name" name="name" placeholder="Name">

            <label for="new-feed-url">Feed URL</label>
            <input type="text" id="new-feed-url" name="url" placeholder="URL">

            <label for="new-feed-category">Category</label>
            <select name="categoryID" id="new-feed-category">
                <option value="" `)
	if args.CurrentCategoryID == "" {
		qw422016.N().S(`selected`)
	}
	qw422016.N().S(`>(no category)</option>
                `)
	for _, cat := range args.Categories {
		qw422016.N().S(`
                <option value="`)
		qw422016.E().S(cat.ID)
		qw422016.N().S(`" `)
		if args.CurrentCategoryID == cat.ID {
			qw422016.N().S(`selected`)
		}
		qw422016.N().S(`>`)
		qw422016.E().S(cat.Name)
		qw422016.N().S(`</option>
                `)
	}
	qw422016.N().S(`
            </select>
        </div>
        <div class="flex-horizontal pt">
            <button class="button green" hx-post="`)
	qw422016.E().S(urls.NewFeedItem)
	qw422016.N().S(`" hx-target="#modal-target">Submit</button>
            <button class="button red" hx-get="`)
	qw422016.E().S(urls.CancelModal)
	qw422016.N().S(`" hx-target="#modal-target">Cancel</button>
        </div>
    </form>
    `)
	components.StreamEndModal(qw422016)
	qw422016.N().S(`
`)
}

func WriteFragmentNewFeed(qq422016 qtio422016.Writer, args *FragmentNewFeedArgs) {
	qw422016 := qt422016.AcquireWriter(qq422016)
	StreamFragmentNewFeed(qw422016, args)
	qt422016.ReleaseWriter(qw422016)
}

func FragmentNewFeed(args *FragmentNewFeedArgs) string {
	qb422016 := qt422016.AcquireByteBuffer()
	WriteFragmentNewFeed(qb422016, args)
	qs422016 := string(qb422016.B)
	qt422016.ReleaseByteBuffer(qb422016)
	return qs422016
}

type FragmentEditFeedArgs struct {
	Feed              *db.Feed
	CurrentCategoryID string
	Categories        []*db.Category
}

func StreamFragmentEditFeed(qw422016 *qt422016.Writer, args *FragmentEditFeedArgs) {
	qw422016.N().S(`
    `)
	components.StreamBeginModal(qw422016)
	qw422016.N().S(`
    <h2><i class="bi bi-pencil-square"></i> Edit Feed</h2>
    <form>
        <div class="form-grid">
            <label for="edit-feed-name">Feed name</label>
            <input type="text" id="edit-feed-name" name="name" placeholder="Name" value="`)
	qw422016.E().S(args.Feed.Name)
	qw422016.N().S(`">

            <label for="edit-feed-url">Feed URL</label>
            <input type="text" id="edit-feed-url" name="url" placeholder="URL" value="`)
	qw422016.E().S(args.Feed.URL)
	qw422016.N().S(`">

            <label for="edit-feed-category">Category</label>
            <select name="categoryID" id="edit-feed-category">
                <option value="" `)
	if args.CurrentCategoryID == "" {
		qw422016.N().S(`selected`)
	}
	qw422016.N().S(`>(no category)</option>
                `)
	for _, cat := range args.Categories {
		qw422016.N().S(`
                    <option value="`)
		qw422016.E().S(cat.ID)
		qw422016.N().S(`" `)
		if args.CurrentCategoryID == cat.ID {
			qw422016.N().S(`selected`)
		}
		qw422016.N().S(`>`)
		qw422016.E().S(cat.Name)
		qw422016.N().S(`</option>
                `)
	}
	qw422016.N().S(`
            </select>
        </div>
        <div class="flex-horizontal pt">
            <button class="button green" hx-put="`)
	qw422016.E().S(urls.Expand(urls.EditFeedItem, args.Feed.ID))
	if args.CurrentCategoryID != "" {
		qw422016.N().S(`?category=`)
		qw422016.E().S(args.CurrentCategoryID)
	}
	qw422016.N().S(`" hx-target="#modal-target">Submit</button>
            <button class="button red" hx-get="`)
	qw422016.E().S(urls.CancelModal)
	qw422016.N().S(`" hx-target="#modal-target">Cancel</button>
        </div>
    </form>
    `)
	components.StreamEndModal(qw422016)
	qw422016.N().S(`
`)
}

func WriteFragmentEditFeed(qq422016 qtio422016.Writer, args *FragmentEditFeedArgs) {
	qw422016 := qt422016.AcquireWriter(qq422016)
	StreamFragmentEditFeed(qw422016, args)
	qt422016.ReleaseWriter(qw422016)
}

func FragmentEditFeed(args *FragmentEditFeedArgs) string {
	qb422016 := qt422016.AcquireByteBuffer()
	WriteFragmentEditFeed(qb422016, args)
	qs422016 := string(qb422016.B)
	qt422016.ReleaseByteBuffer(qb422016)
	return qs422016
}
