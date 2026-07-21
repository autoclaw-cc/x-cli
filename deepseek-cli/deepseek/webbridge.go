package deepseek

import "fmt"

type AnswerSnapshot struct {
	Count  int    `json:"count"`
	Latest string `json:"latest"`
}

type SubmitResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

type ModeResult struct {
	OK      bool     `json:"ok"`
	Error   string   `json:"error"`
	Enabled []string `json:"enabled"`
	Clicked []string `json:"clicked"`
}

type UploadTargetResult struct {
	OK       bool   `json:"ok"`
	Error    string `json:"error"`
	Selector string `json:"selector"`
}

type NewChatResult struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error"`
	Started bool   `json:"started"`
}

const SnapshotScript = `
(() => {
  const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
  const txt=s=>(s||'').replace(/\s+/g,' ').trim();
  const body=txt(document.body?.innerText||'');
  const inputs=[...document.querySelectorAll('textarea,[contenteditable=true],input')]
    .filter(visible).slice(0,12).map(e=>({
      tag:e.tagName,
      role:e.getAttribute('role')||'',
      type:e.getAttribute('type')||'',
      aria:e.getAttribute('aria-label')||'',
      placeholder:e.getAttribute('placeholder')||'',
      text:txt(e.innerText||e.textContent||e.value).slice(0,120),
      disabled:!!e.disabled,
      cls:String(e.className).slice(0,120)
    }));
  const re=/DeepThink|R1|深度思考|联网|智能搜索|Search|搜索|上传|Attach|发送|Send|新对话|New chat|设置|Settings|模型|model|文件|识图|vision/i;
  const controls=[...document.querySelectorAll('button,[role=button],[role=radio],a,div,span')]
    .filter(e=>visible(e)&&re.test(txt(e.innerText||e.textContent)+' '+(e.getAttribute('aria-label')||'')))
    .slice(0,40).map(e=>({
      tag:e.tagName,
      role:e.getAttribute('role')||'',
      aria:e.getAttribute('aria-label')||'',
      text:txt(e.innerText||e.textContent).slice(0,120),
      disabled:!!e.disabled,
      cls:String(e.className).slice(0,120)
    }));
  return {href:location.href,title:document.title,bodyText:body.slice(0,600),hasPromptInput:inputs.length>0,inputs,controls};
})()
`

func SetModesScript(deepthink, search bool) string {
	return fmt.Sprintf(`
(() => {
  function deepseekSetModes() {
    const desired={deepthink:%t,web_search:%t};
    const txt=s=>(s||'').replace(/\s+/g,' ').trim();
    const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
    const controls=[...document.querySelectorAll('button,[role=button],[role=checkbox],[aria-pressed],div,span')]
      .filter(e=>visible(e));
    const specs={
      deepthink:[/深度思考/i,/DeepThink/i,/Deep Think/i,/\bR1\b/i],
      web_search:[/智能搜索/i,/联网搜索/i,/联网/i,/\bSearch\b/i,/搜索/i]
    };
    const active=e=>{
      const attrs=[e.getAttribute('aria-pressed'),e.getAttribute('aria-checked'),e.getAttribute('data-state')].join(' ').toLowerCase();
      const cls=String(e.className||'').toLowerCase();
      const label=(txt(e.innerText||e.textContent)+' '+(e.getAttribute('aria-label')||'')).toLowerCase();
      return /\btrue\b|checked|selected|active|on/.test(attrs+' '+cls) || /已开启|开启中|enabled|active/.test(label);
    };
    const actionable=e=>e.closest('button,[role=button],[role=checkbox]')||e;
    const enabled=[],clicked=[],missing=[];
    for (const [key,want] of Object.entries(desired)) {
      if(!want) continue;
      const found=controls.find(e=>specs[key].some(re=>re.test(txt(e.innerText||e.textContent)+' '+(e.getAttribute('aria-label')||''))));
      if(!found){missing.push(key);continue;}
      const target=actionable(found);
      if(!active(target)){target.click();clicked.push(key);}
      enabled.push(key);
    }
    if(missing.length) return {ok:false,error:'mode_control_missing:'+missing.join(','),enabled,clicked};
    return {ok:true,enabled,clicked};
  }
  return deepseekSetModes();
})()
`, deepthink, search)
}

const PrepareUploadScript = `
(async () => {
  function deepseekPrepareUpload() {
    const txt=s=>(s||'').replace(/\s+/g,' ').trim();
    const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
    const selector='input[type=file]';
    let input=document.querySelector(selector);
    if(input) return {ok:true,selector};
    const control=[...document.querySelectorAll('button,[role=button],div,span')]
      .filter(visible)
      .find(e=>/上传|文件|Attach|Upload|File/i.test(txt(e.innerText||e.textContent)+' '+(e.getAttribute('aria-label')||'')));
    if(control) (control.closest('button,[role=button]')||control).click();
    return {ok:!!document.querySelector(selector),selector,error:document.querySelector(selector)?'':'file_input_missing'};
  }
  let result=deepseekPrepareUpload();
  if(result.ok) return result;
  await new Promise(resolve=>setTimeout(resolve,250));
  return deepseekPrepareUpload();
})()
`

const AnswerSnapshotScript = `
(() => {
  const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
  const txt=s=>(s||'').replace(/\s+/g,' ').trim();
  const answers=[...document.querySelectorAll('.ds-assistant-message-main-content')]
    .filter(visible).map(e=>txt(e.innerText||e.textContent)).filter(Boolean);
  return {count:answers.length,latest:answers[answers.length-1]||''};
})()
`

const NewChatScript = `
(() => {
  function deepseekNewChat() {
    const txt=s=>(s||'').replace(/\s+/g,' ').trim();
    const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
    const target=[...document.querySelectorAll('button,[role=button],a,div,span')]
      .filter(visible)
      .find(e=>/新对话|新建对话|New chat|New conversation/i.test(txt(e.innerText||e.textContent)+' '+(e.getAttribute('aria-label')||'')));
    if(!target) return {ok:false,error:'new_chat_control_missing',started:false};
    (target.closest('button,[role=button],a')||target).click();
    return {ok:true,started:true};
  }
  return deepseekNewChat();
})()
`

const SubmitPromptScript = `
(() => {
  function deepseekSubmitPrompt() {
    const textarea=document.querySelector('textarea');
    if(!textarea) return {ok:false,error:'textarea_missing'};
    const tr=textarea.getBoundingClientRect();
    const candidates=[...document.querySelectorAll('[role=button],button')]
      .map(e=>({e,r:e.getBoundingClientRect(),text:(e.innerText||e.textContent||'').trim()}))
      .filter(x=>x.r.width>20&&x.r.height>20&&x.r.y>=tr.y&&x.r.x>tr.x+tr.width*0.65)
      .sort((a,b)=>b.r.x-a.r.x);
    const target=candidates[0];
    if(!target) return {ok:false,error:'send_button_missing'};
    target.e.click();
    return {ok:true};
  }
  return deepseekSubmitPrompt();
})()
`
