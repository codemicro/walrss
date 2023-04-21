// Code generated by qtc from "settings.qtpl.html". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

package neoviews

import "github.com/codemicro/walrss/walrss/internal/http/neoviews/internal/components"

import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

func StreamSettingsPage(qw422016 *qt422016.Writer) {
	qw422016.N().S(`
<!DOCTYPE html>
<html lang="en">
`)
	components.StreamHead(qw422016, "Settings")
	qw422016.N().S(`
<body>
`)
	components.StreamNavigation(qw422016, menuItems, menuItemSettings)
	qw422016.N().S(`

`)
	components.StreamBeginMain(qw422016)
	qw422016.N().S(`

<div class="container">
    <h1 class="title"><i class="bi bi-gear-fill"></i> Settings</h1>

    <div class="card">
        <h2>Delivery Controls</h2>

        <form action="">
            <input type="checkbox" name="" id="f"><label for="f">Enable digest delivery</label>
        </form>

        <div class="pt"></div>

        <form action="">
            Deliver my digests
            <select name="day" id="day-selector">
                <option value="cheeze">Cheese</option>
            </select>
            at
            <select name="time" id="time-selector">
                <option value="5am">5am</option>
            </select>
            UTC
            <button class="button inline">Save</button>
        </form>
    </div>

    <div class="card">
        <h2>Import and Export</h2>
        <p>You can import and export your feeds as OPML files using the buttons below. This will allow you to move your feed collection between different feed readers and processors.</p>
        <p><span class="warning-text">WARNING:</span> you have at least one JSON feed in your feed collection. These will be included in the OPML export, but may cause issues when being imported into other feed readers.</p>
    </div>
</div>

`)
	components.StreamEndMain(qw422016)
	qw422016.N().S(`

</body>
</html>
`)
}

func WriteSettingsPage(qq422016 qtio422016.Writer) {
	qw422016 := qt422016.AcquireWriter(qq422016)
	StreamSettingsPage(qw422016)
	qt422016.ReleaseWriter(qw422016)
}

func SettingsPage() string {
	qb422016 := qt422016.AcquireByteBuffer()
	WriteSettingsPage(qb422016)
	qs422016 := string(qb422016.B)
	qt422016.ReleaseByteBuffer(qb422016)
	return qs422016
}
