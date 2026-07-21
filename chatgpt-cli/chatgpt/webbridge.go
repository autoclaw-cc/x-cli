package chatgpt

import "fmt"

const ReadyScript = `(()=>{
  function chatgptPageReady(){
    const home=location.origin==='https://chatgpt.com'&&location.pathname==='/';
    const prompt=document.querySelector('#prompt-textarea');
    const form=prompt?.closest('form');
    const hydrated=!!form&&form.querySelectorAll('button').length>=3;
    const login=!!document.querySelector('[data-testid="login-button"],[data-testid="signup-button"]');
    return {ready:home&&((!!prompt&&hydrated)||login)};
  }
  return JSON.stringify(chatgptPageReady());
})()`

const SnapshotScript = `(async()=>{
  function chatgptSnapshot(){
    return true;
  }
  chatgptSnapshot();
  const prompt=document.querySelector('#prompt-textarea');
  const loginControls=!!document.querySelector('[data-testid="login-button"],[data-testid="signup-button"]');
  const labels=[];
  const wanted=/^(创建图片|网页搜索|深度研究|Create image|Web search|Deep research)$/i;
  const readLabels=()=>{
    for(const e of document.querySelectorAll('span,div')){
      if(e.closest('nav,aside')) continue;
      const rect=e.getBoundingClientRect();
      if(!rect.width||!rect.height) continue;
      const own=[...e.childNodes].filter(n=>n.nodeType===Node.TEXT_NODE).map(n=>n.textContent.trim()).join(' ').trim();
      if(wanted.test(own)&&!labels.includes(own)) labels.push(own);
    }
  };
  readLabels();
  let opened=false;
  if(prompt&&!loginControls&&labels.length===0){
    const plus=document.querySelector('[data-testid="composer-plus-btn"]');
    if(plus){
      plus.click();
      opened=true;
      const deadline=Date.now()+6000;
      while(Date.now()<deadline&&labels.length<3){
        await new Promise(resolve=>setTimeout(resolve,100));
        readLabels();
      }
    }
  }
  if(opened){
    document.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape',bubbles:true}));
    await new Promise(resolve=>setTimeout(resolve,50));
  }
  return JSON.stringify({
    href:location.href,
    locale:document.documentElement.lang||navigator.language||'',
    hasPrompt:!!prompt,
    loginControls,
    toolLabels:[...new Set(labels)]
  });
})()`

type Citation struct {
	URL   string `json:"url"`
	Label string `json:"label,omitempty"`
}

type AnswerSnapshot struct {
	Count     int        `json:"count"`
	Latest    string     `json:"latest"`
	Streaming bool       `json:"streaming"`
	Citations []Citation `json:"citations"`
}

type ModeResult struct {
	OK    bool   `json:"ok"`
	Mode  string `json:"mode"`
	Error string `json:"error"`
}

type SubmitResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func SelectModeScript(mode string) string {
	return fmt.Sprintf(`(async()=>{
  function chatgptSelectMode(){ return true; }
  chatgptSelectMode();
  const mode=%q;
  const patterns=mode==='deep_research'
    ? [/^深度研究$/i,/^Deep research$/i]
    : [/^网页搜索$/i,/^Web search$/i];
  const plus=document.querySelector('[data-testid="composer-plus-btn"]');
  if(!plus) return JSON.stringify({ok:false,mode,error:'tool_menu_missing'});
  plus.click();
  const deadline=Date.now()+8000;
  let item=null;
  while(Date.now()<deadline&&!item){
    for(const e of document.querySelectorAll('span,div')){
      if(e.closest('nav,aside')) continue;
      const rect=e.getBoundingClientRect();
      if(!rect.width||!rect.height) continue;
      const own=[...e.childNodes].filter(n=>n.nodeType===Node.TEXT_NODE).map(n=>n.textContent.trim()).join(' ').trim();
      if(patterns.some(re=>re.test(own))){ item=e.closest('[tabindex="0"]'); break; }
    }
    if(!item) await new Promise(resolve=>setTimeout(resolve,100));
  }
  if(!item) return JSON.stringify({ok:false,mode,error:'mode_not_available'});
  item.click();
  await new Promise(resolve=>setTimeout(resolve,250));
  const form=document.querySelector('#prompt-textarea')?.closest('form');
  const text=(form?.innerText||'').toLowerCase();
  const active=mode==='deep_research'
    ? (text.includes('深度研究')||text.includes('deep research'))
    : (text.includes('网页搜索')||text.includes('web search'));
  return JSON.stringify(active?{ok:true,mode}:{ok:false,mode,error:'mode_not_active'});
})()`, mode)
}

const SubmitPromptScript = `(()=>{
  function chatgptSubmitPrompt(){
    const button=document.querySelector('[data-testid="send-button"],#composer-submit-button');
    if(!button) return {ok:false,error:'send_button_missing'};
    if(button.disabled) return {ok:false,error:'send_button_not_ready'};
    button.click();
    return {ok:true};
  }
  return JSON.stringify(chatgptSubmitPrompt());
})()`

const AnswerSnapshotScript = `(()=>{
  function chatgptAnswerSnapshot(){
    const messages=[...document.querySelectorAll('[data-message-author-role="assistant"]')];
    const latest=messages.at(-1);
    const citations=[];
    if(latest){
      for(const anchor of latest.querySelectorAll('a[href]')){
        const url=anchor.href||'';
        if(!/^https?:/i.test(url)) continue;
        citations.push({url,label:(anchor.innerText||anchor.getAttribute('aria-label')||'').trim().slice(0,160)});
      }
    }
    const unique=[];
    const seen=new Set();
    for(const citation of citations){if(!seen.has(citation.url)){seen.add(citation.url);unique.push(citation);}}
    return {
      count:messages.length,
      latest:(latest?.innerText||'').trim(),
      streaming:!!document.querySelector('[data-testid="stop-button"],[aria-label*="停止回答"],[aria-label*="停止生成"],[aria-label*="Stop"]'),
      citations:unique
    };
  }
  return JSON.stringify(chatgptAnswerSnapshot());
})()`
