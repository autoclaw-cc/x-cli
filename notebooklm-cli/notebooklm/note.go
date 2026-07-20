package notebooklm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type NoteResult struct {
	Title          string `json:"title"`
	BodyCharacters int    `json:"body_characters"`
}

type NoteSummary struct {
	Title string `json:"title"`
}

type NoteListResult struct {
	Notes []NoteSummary `json:"notes"`
}

type NoteSourceResult struct {
	Title       string `json:"title"`
	SourceCount int    `json:"source_count"`
}

func CreateNote(ctx context.Context, bridge Bridge, notebookURL, title, body string, timeout time.Duration) (*NoteResult, error) {
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("note title is empty")
	}
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("note body is empty")
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	baseline, err := openStudioNotes(ctx, bridge, title, deadline)
	if err != nil {
		return nil, err
	}

	const tagAdd = `(() => {const b=[...document.querySelectorAll('button')].find(e=>/添加笔记|Add note/i.test((e.textContent||'').trim()));if(!b)return {ok:false};b.id='notebooklm-add-note';b.click();return {ok:true};})()`
	var tagged struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(tagAdd, &tagged); err != nil || !tagged.OK {
		if err == nil {
			err = fmt.Errorf("add note control not found")
		}
		return nil, err
	}
	if err := waitNoteEditor(ctx, bridge, deadline, "", ""); err != nil {
		return nil, err
	}

	encodedTitle, _ := json.Marshal(title)
	setTitle := fmt.Sprintf(`(() => {const i=document.querySelector('.note-header__editable-title');if(!i)return {ok:false,title:''};const value=%s;const set=Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set;set.call(i,value);i.dispatchEvent(new InputEvent('input',{bubbles:true,inputType:'insertText',data:value}));i.dispatchEvent(new Event('change',{bubbles:true}));i.blur();return {ok:i.value===value,title:i.value};})()`, encodedTitle)
	var renamed struct {
		OK    bool   `json:"ok"`
		Title string `json:"title"`
	}
	if err := bridge.EvaluateValue(setTitle, &renamed); err != nil || !renamed.OK || renamed.Title != title {
		if err == nil {
			err = fmt.Errorf("note title verification failed")
		}
		return nil, err
	}
	if err := bridge.Fill(".ProseMirror", body); err != nil {
		return nil, fmt.Errorf("fill note body: %w", err)
	}
	if err := waitNoteEditor(ctx, bridge, deadline, title, body); err != nil {
		return nil, err
	}
	if err := closeNoteEditor(bridge); err != nil {
		return nil, err
	}

	encodedTitleString := string(encodedTitle)
	waitPersisted := fmt.Sprintf(`(() => {const notes=[...document.querySelectorAll('artifact-library-note')];const matches=notes.filter(e=>(e.querySelector('.artifact-title')?.textContent||'').trim()===%s);return {ready:notes.length>%d&&matches.length===1,noteCount:notes.length};})()`, encodedTitleString, baseline)
	if err := waitNoteReady(ctx, bridge, waitPersisted, deadline, "timeout: note did not appear in Studio"); err != nil {
		return nil, err
	}
	return &NoteResult{Title: title, BodyCharacters: len([]rune(body))}, nil
}

func ListNotes(ctx context.Context, bridge Bridge, notebookURL string, timeout time.Duration) (*NoteListResult, error) {
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	if _, err := openStudioNotes(ctx, bridge, "", deadline); err != nil {
		return nil, err
	}
	const inspect = `(() => {const visible=e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length);const loading=[...document.querySelectorAll('mat-spinner,mat-progress-spinner,[role=progressbar]')].some(visible);return {loading,notes:[...document.querySelectorAll('artifact-library-note')].map(e=>({title:(e.querySelector('.artifact-title')?.textContent||'').trim()})).filter(e=>e.title)}})()`
	lastKey := ""
	stable := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state struct {
			Loading bool          `json:"loading"`
			Notes   []NoteSummary `json:"notes"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && !state.Loading {
			titles := make([]string, 0, len(state.Notes))
			for _, note := range state.Notes {
				titles = append(titles, note.Title)
			}
			key := strings.Join(titles, "\x00")
			if key == lastKey {
				stable++
			} else {
				lastKey = key
				stable = 1
			}
			if stable >= 4 {
				if state.Notes == nil {
					state.Notes = []NoteSummary{}
				}
				return &NoteListResult{Notes: state.Notes}, nil
			}
		} else {
			stable = 0
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: note list did not stabilize")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func ConvertNoteToSource(ctx context.Context, bridge Bridge, notebookURL, title string, timeout time.Duration) (*NoteSourceResult, error) {
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("note title is empty")
	}
	if err := openOwnedNotebook(ctx, bridge, notebookURL, timeout); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	baseline, err := openSourcesAndCount(ctx, bridge, deadline)
	if err != nil {
		return nil, err
	}
	if _, err := openStudioNotes(ctx, bridge, "", deadline); err != nil {
		return nil, err
	}
	encodedTitle, _ := json.Marshal(title)
	tagNote := fmt.Sprintf(`(() => {const notes=[...document.querySelectorAll('artifact-library-note')];const matches=notes.filter(e=>(e.querySelector('.artifact-title')?.textContent||'').trim()===%s);const card=matches[0];const b=card?.querySelector('button.artifact-stretched-button')||card?.querySelector('.artifact-item-button');if(matches.length!==1||!b)return {ok:false,matches:matches.length};b.id='notebooklm-convert-note-target';b.click();return {ok:true,matches:1};})()`, encodedTitle)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var state struct {
			OK      bool `json:"ok"`
			Matches int  `json:"matches"`
		}
		if err := bridge.EvaluateValue(tagNote, &state); err == nil {
			if state.Matches > 1 {
				return nil, fmt.Errorf("note title is not unique")
			}
			if state.OK {
				break
			}
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: note was not found")
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err := waitNoteEditor(ctx, bridge, deadline, title, ""); err != nil {
		return nil, err
	}

	const tagConvert = `(() => {const visible=e=>!!(e.offsetWidth||e.offsetHeight||e.getClientRects().length);const b=[...document.querySelectorAll('button')].filter(visible).find(e=>/转换为来源|Convert to source/i.test((e.textContent||'').trim()));if(!b)return {ok:false};b.id='notebooklm-convert-note';return {ok:true};})()`
	var convert struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(tagConvert, &convert); err != nil || !convert.OK {
		if err == nil {
			err = fmt.Errorf("convert note control not found")
		}
		return nil, err
	}
	if err := bridge.MouseClick("#notebooklm-convert-note"); err != nil {
		return nil, fmt.Errorf("convert note to source: %w", err)
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		sourceCount, err := openSourcesAndCount(ctx, bridge, deadline)
		if err == nil && sourceCount > baseline {
			return &NoteSourceResult{Title: title, SourceCount: sourceCount}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout: converted note did not become a source")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func openStudioNotes(ctx context.Context, bridge Bridge, uniqueTitle string, deadline time.Time) (int, error) {
	const tagStudio = `(() => {const e=[...document.querySelectorAll('[role=tab]')].find(x=>(x.textContent||'').trim()==='Studio');if(!e)return {ok:false};e.id='notebooklm-notes-studio-tab';e.click();return {ok:true};})()`
	var tagged struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(tagStudio, &tagged); err != nil || !tagged.OK {
		if err == nil {
			err = fmt.Errorf("Studio tab not found")
		}
		return 0, err
	}
	encodedTitle, _ := json.Marshal(uniqueTitle)
	inspect := fmt.Sprintf(`(() => {const tab=document.querySelector('#notebooklm-notes-studio-tab');if(tab&&tab.getAttribute('aria-selected')!=='true'){tab.click();return {ready:false,noteCount:0,titleExists:false};}const notes=[...document.querySelectorAll('artifact-library-note')];const title=%s;return {ready:tab?.getAttribute('aria-selected')==='true'&&!![...document.querySelectorAll('button')].find(e=>/添加笔记|Add note/i.test((e.textContent||'').trim())),noteCount:notes.length,titleExists:!!title&&notes.some(e=>(e.querySelector('.artifact-title')?.textContent||'').trim()===title)};})()`, encodedTitle)
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		var state struct {
			Ready       bool `json:"ready"`
			NoteCount   int  `json:"noteCount"`
			TitleExists bool `json:"titleExists"`
		}
		if err := bridge.EvaluateValue(inspect, &state); err == nil && state.Ready {
			if state.TitleExists {
				return 0, fmt.Errorf("note title already exists")
			}
			return state.NoteCount, nil
		}
		if time.Now().After(deadline) {
			return 0, fmt.Errorf("timeout: Studio notes did not become ready")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func waitNoteEditor(ctx context.Context, bridge Bridge, deadline time.Time, title, body string) error {
	encodedTitle, _ := json.Marshal(title)
	encodedBody, _ := json.Marshal(body)
	inspect := fmt.Sprintf(`(() => {const i=document.querySelector('.note-header__editable-title');const editor=document.querySelector('.ProseMirror');const title=%s;const body=%s;return {ready:!!i&&!!editor&&(!title||i.value===title)&&(!body||(editor.innerText||'')===body)};})()`, encodedTitle, encodedBody)
	return waitNoteReady(ctx, bridge, inspect, deadline, "timeout: note editor did not become ready")
}

func closeNoteEditor(bridge Bridge) error {
	const tagClose = `(() => {const b=[...document.querySelectorAll('button')].find(e=>/关闭笔记|Close note/i.test(e.getAttribute('aria-label')||''));if(!b)return {ok:false};b.id='notebooklm-close-note';b.click();return {ok:true};})()`
	var tagged struct {
		OK bool `json:"ok"`
	}
	if err := bridge.EvaluateValue(tagClose, &tagged); err != nil || !tagged.OK {
		if err == nil {
			err = fmt.Errorf("close note control not found")
		}
		return err
	}
	return nil
}

func waitNoteReady(ctx context.Context, bridge Bridge, script string, deadline time.Time, timeoutMessage string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var state struct {
			Ready bool `json:"ready"`
		}
		if err := bridge.EvaluateValue(script, &state); err == nil && state.Ready {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s", timeoutMessage)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
