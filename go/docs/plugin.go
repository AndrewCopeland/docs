package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/blang/semver"
	_ "github.com/concourse/docs/go/chromastyle"

	"github.com/vito/booklit"
	"github.com/vito/booklit/ast"
	"github.com/vito/booklit/chroma"
	"github.com/vito/booklit/stages"
)

func init() {
	booklit.RegisterPlugin("concourse-docs", NewPlugin)
}

type Plugin struct {
	section *booklit.Section
	chroma  chroma.Plugin

	definitionContext []string
	noteIdx           int
}

func NewPlugin(section *booklit.Section) booklit.Plugin {
	return &Plugin{
		section: section,
		chroma:  chroma.NewPlugin(section).(chroma.Plugin),
	}
}

func (p Plugin) FontAwesome(class string) booklit.Content {
	return booklit.Styled{
		Style:   "font-awesome",
		Content: booklit.String(class),
	}
}

func (p Plugin) Codeblock(language string, code booklit.Content) (booklit.Content, error) {
	return p.chroma.Syntax(language, code, "concourseci")
}

func (p Plugin) InlineHeader(content booklit.Content) booklit.Content {
	return booklit.Styled{
		Style:   "inline-header",
		Content: content,
	}
}

func (p Plugin) SplashIntro(intro, blurb booklit.Content) booklit.Content {
	return booklit.Styled{
		Style: "splash-intro",
		Block: true,

		Content: blurb,

		Partials: booklit.Partials{
			"Intro": intro,
		},
	}
}

func (p Plugin) QuickStart(content booklit.Content) booklit.Content {
	return booklit.Styled{
		Style:   "quick-start",
		Block:   true,
		Content: content,
	}
}

func (p Plugin) SplashExample(title booklit.Content, content booklit.Content, example booklit.Content) booklit.Content {
	return booklit.Styled{
		Style: "splash-example",
		Block: true,

		Content: content,

		Partials: booklit.Partials{
			"Title":   title,
			"Example": example,
		},
	}
}

func (p Plugin) ExamplePipeline() booklit.Content {
	return booklit.Styled{
		Style:   "example-pipeline",
		Block:   true,
		Content: booklit.Empty,
	}
}

func (p Plugin) TitledCodeblock(title booklit.Content, language string, code booklit.Content) (booklit.Content, error) {
	codeblock, err := p.Codeblock(language, code)
	if err != nil {
		return nil, err
	}

	return booklit.Styled{
		Style: "titled-codeblock",
		Block: true,

		Content: codeblock,

		Partials: booklit.Partials{
			"Title": booklit.Styled{
				Style:   booklit.StyleVerbatim,
				Content: title,
			},
		},
	}, nil
}

func (p Plugin) Warn(content booklit.Content) booklit.Content {
	return booklit.Styled{
		Style:   "warning",
		Content: content,
	}
}

func (p Plugin) Frame(src booklit.Content) booklit.Content {
	return booklit.Styled{
		Style:   "frame",
		Content: booklit.Empty,
		Partials: booklit.Partials{
			"Src": src,
		},
	}
}

func (p *Plugin) DefineAttribute(attribute string, contentNode ast.Node, tags ...string) (booklit.Content, error) {
	attrSplit := strings.SplitN(attribute, ":", 2)

	attrName := attrSplit[0]
	if len(tags) == 0 {
		tags = []string{attrName}
	}

	oldCtx := p.definitionContext
	p.definitionContext = append(p.definitionContext, attrName)

	stage := &stages.Evaluate{
		Section: p.section,
	}

	err := contentNode.Visit(stage)
	if err != nil {
		return nil, err
	}

	content := stage.Result

	display := booklit.Styled{
		Style: booklit.StyleBold,
		Content: booklit.Styled{
			Style:   booklit.StyleVerbatim,
			Content: booklit.String(strings.Join(p.definitionContext, ".")),
		},
	}

	p.definitionContext = oldCtx

	targets := booklit.Sequence{}
	for _, t := range tags {
		targets = append(targets, booklit.Target{
			TagName: t,
			Title:   display,
			Content: content,
		})
	}

	return NoIndex{
		booklit.Styled{
			Style:   "definition",
			Content: content,
			Partials: booklit.Partials{
				"Targets": targets,
				"Thumb": booklit.Styled{
					Style: booklit.StyleVerbatim,
					Content: booklit.Preformatted{
						booklit.Sequence{
							&booklit.Reference{
								TagName: tags[0],
								Content: booklit.Styled{
									Style:   booklit.StyleBold,
									Content: booklit.String(attrName),
								},
							},
							booklit.String(":" + attrSplit[1]),
						},
					},
				},
			},
		},
	}, nil
}

func (p Plugin) DefineMetric(metric string, content booklit.Content) booklit.Content {
	return NoIndex{
		booklit.Styled{
			Style:   "definition",
			Content: content,
			Partials: booklit.Partials{
				"Targets": booklit.Target{
					TagName: metric,
					Title:   booklit.String(metric),
					Content: content,
				},
				"Thumb": booklit.Styled{
					Style:   booklit.StyleVerbatim,
					Content: booklit.Preformatted{booklit.String(metric)},
				},
			},
		},
	}
}

func (p Plugin) DefineTable(table string, content booklit.Content) booklit.Content {
	tagName := table + "-table"

	return NoIndex{
		booklit.Styled{
			Style:   "definition",
			Content: content,
			Partials: booklit.Partials{
				"Targets": booklit.Target{
					TagName: tagName,
					Title: booklit.Styled{
						Style: booklit.StyleVerbatim,
						Content: booklit.Styled{
							Style:   booklit.StyleBold,
							Content: booklit.String(table),
						},
					},
					Content: content,
				},
				"Thumb": booklit.Styled{
					Style: booklit.StyleVerbatim,
					Content: booklit.Preformatted{
						&booklit.Reference{
							TagName: tagName,
							Content: booklit.Styled{
								Style:   booklit.StyleBold,
								Content: booklit.String(table),
							},
						},
					},
				},
			},
		},
	}
}

func (p Plugin) ReferenceColumn(table string, column string) booklit.Content {
	return &booklit.Reference{
		TagName: table + "-table",
		Content: booklit.Styled{
			Style: booklit.StyleVerbatim,
			Content: booklit.Sequence{
				booklit.Styled{
					Style:   booklit.StyleBold,
					Content: booklit.String(table),
				},
				booklit.String(" (" + column + ")"),
			},
		},
	}
}

func (p Plugin) Boshprop(job string, target string) booklit.Content {
	return booklit.Link{
		Target: fmt.Sprintf("https://bosh.io/jobs/%s?source=github.com/concourse/concourse-bosh-release#p=%s", job, target),
		Content: booklit.Styled{
			Style:   booklit.StyleVerbatim,
			Content: booklit.String(target),
		},
	}
}

func (p Plugin) Ghuser(user string) booklit.Content {
	return booklit.Styled{
		Style:   "github-user-link",
		Content: booklit.String(user),
	}
}

func (p Plugin) Ghpr(number string, optionalRepo ...string) booklit.Content {
	repo := "concourse"
	if len(optionalRepo) > 0 {
		repo = optionalRepo[0]
	}

	return booklit.Styled{
		Style:   "github-pr-link",
		Content: booklit.String(number),
		Partials: booklit.Partials{
			"Owner": booklit.String("concourse"),
			"Repo":  booklit.String(repo),
		},
	}
}

func (p Plugin) Ghissue(number string, optionalRepo ...string) booklit.Content {
	repo := "concourse"
	if len(optionalRepo) > 0 {
		repo = optionalRepo[0]
	}

	return booklit.Styled{
		Style:   "github-issue-link",
		Content: booklit.String(number),
		Partials: booklit.Partials{
			"Owner": booklit.String("concourse"),
			"Repo":  booklit.String(repo),
		},
	}
}

func (p Plugin) Resource(name string) booklit.Content {
	return booklit.Link{
		Target: fmt.Sprintf("https://github.com/concourse/%s-resource", name),
		Content: booklit.Sequence{
			booklit.Styled{
				Style:   booklit.StyleVerbatim,
				Content: booklit.String(name),
			},
			booklit.String(" resource"),
		},
	}
}

func (p Plugin) TutorialImage(path string) booklit.Content {
	return booklit.Styled{
		Style: "classed-image",
		Content: booklit.Image{
			Path:        path,
			Description: "tutorial image",
		},
		Partials: booklit.Partials{"Class": booklit.String("tutorial-image")},
	}
}

func (p Plugin) LiterateSegment(parasAndFinalCode ...booklit.Content) (booklit.Content, error) {
	if len(parasAndFinalCode) == 0 {
		return nil, fmt.Errorf("no paragraphs or code given")
	}

	paras := parasAndFinalCode[0 : len(parasAndFinalCode)-1]
	code := parasAndFinalCode[len(parasAndFinalCode)-1]

	if len(paras) == 0 {
		paras = []booklit.Content{code}
		code = booklit.Empty
	}

	return booklit.Styled{
		Style:   "literate-segment",
		Content: booklit.Sequence(paras),
		Partials: booklit.Partials{
			"Code": code,
		},
	}, nil
}

func (p Plugin) PipelineImage(path string) booklit.Content {
	return booklit.Styled{
		Style: "pipeline-image",
		Content: booklit.Image{
			Path:        path,
			Description: "pipeline",
		},
	}
}

func (p *Plugin) Note(commaSeparatedTags string, content booklit.Content) booklit.Content {
	tags := strings.Split(commaSeparatedTags, ",")

	p.noteIdx++

	idx := p.noteIdx
	targetTag := fmt.Sprintf("%s-note-%d", p.section.PrimaryTag.Name, idx)

	tagNotes := []booklit.Content{}
	for _, t := range tags {
		tagNotes = append(tagNotes, booklit.Styled{
			Style:   "release-note-tag",
			Content: booklit.String(t),
		})
	}

	return booklit.Styled{
		Style:   "release-note",
		Content: content,
		Partials: booklit.Partials{
			"Tags":   booklit.Sequence(tagNotes),
			"Target": booklit.String(targetTag),
		},
	}
}

func (p Plugin) RightSide(title, content booklit.Content) {
	wrappedContent := booklit.Styled{
		Style:   "sidebar-right",
		Content: content,
		Partials: booklit.Partials{
			"Title": title,
		},
	}

	p.section.SetPartial("RightSide", wrappedContent)
}

func (p Plugin) Example(title, content booklit.Content) booklit.Content {
	return booklit.Styled{
		Style:   "example",
		Block:   true,
		Content: content,
		Partials: booklit.Partials{
			"Title": title,
		},
	}
}

func (p Plugin) TrademarkGuidelines(content ...booklit.Content) booklit.Content {
	return booklit.Styled{
		Style: "trademark-guidelines",
		Block: true,
		Content: booklit.List{
			Items: content,
		},
	}
}

var archivedBinariesVersion = semver.MustParse("5.0.0")
var flyBinariesVersion = semver.MustParse("2.2.0")

func (p Plugin) ReleaseVersion(version string) error {
	p.section.Style = "release"

	p.section.SetTitle(booklit.String("v" + version))

	p.section.SetPartial("Version", booklit.String(version))

	v, err := semver.Parse(version)
	if err != nil {
		return err
	}

	downloadURL := "https://github.com/concourse/concourse/releases/download/v" + version

	if v.GTE(archivedBinariesVersion) {
		p.section.SetPartial("ConcourseLinuxURL", booklit.String(downloadURL+"/concourse-"+version+"-linux-amd64.tgz"))
		p.section.SetPartial("ConcourseWindowsURL", booklit.String(downloadURL+"/concourse-"+version+"-windows-amd64.zip"))
		p.section.SetPartial("ConcourseDarwinURL", booklit.String(downloadURL+"/concourse-"+version+"-darwin-amd64.tgz"))

		p.section.SetPartial("HasFlyBinaries", booklit.Empty)
		p.section.SetPartial("FlyLinuxURL", booklit.String(downloadURL+"/fly-"+version+"-linux-amd64.tgz"))
		p.section.SetPartial("FlyWindowsURL", booklit.String(downloadURL+"/fly-"+version+"-windows-amd64.zip"))
		p.section.SetPartial("FlyDarwinURL", booklit.String(downloadURL+"/fly-"+version+"-darwin-amd64.tgz"))
	} else {
		p.section.SetPartial("ConcourseLinuxURL", booklit.String(downloadURL+"/concourse_linux_amd64"))
		p.section.SetPartial("ConcourseWindowsURL", booklit.String(downloadURL+"/concourse_windows_amd64.exe"))
		p.section.SetPartial("ConcourseDarwinURL", booklit.String(downloadURL+"/concourse_darwin_amd64"))

		if v.GTE(flyBinariesVersion) {
			p.section.SetPartial("HasFlyBinaries", booklit.Empty)
			p.section.SetPartial("FlyLinuxURL", booklit.String(downloadURL+"/fly_linux_amd64"))
			p.section.SetPartial("FlyWindowsURL", booklit.String(downloadURL+"/fly_windows_amd64.exe"))
			p.section.SetPartial("FlyDarwinURL", booklit.String(downloadURL+"/fly_darwin_amd64"))
		}
	}

	return nil
}

func (p Plugin) ReleaseGardenLinuxVersion(version string) {
	p.section.SetPartial("GardenLinuxVersion", booklit.String(version))
}

func (p Plugin) ReleaseGardenRuncVersion(version string) {
	p.section.SetPartial("GardenRuncVersion", booklit.String(version))
}

func (p Plugin) ReleaseDate(date string) error {
	t, err := time.Parse("2006-1-2", date)
	if err != nil {
		return err
	}

	p.section.SetPartial("ReleaseDate", booklit.Styled{
		Style:   "release-date",
		Content: booklit.String(t.Format("January 2, 2006")),
	})

	return nil
}

type NoIndex struct {
	booklit.Content
}

func (NoIndex) String() string {
	return ""
}
