package chatgpt

const ReadyScript = `(()=>{
  function chatgptPageReady(){
    const home=location.origin==='https://chatgpt.com'&&location.pathname==='/';
    const prompt=!!document.querySelector('#prompt-textarea');
    const login=!!document.querySelector('[data-testid="login-button"],[data-testid="signup-button"]');
    return {ready:home&&(prompt||login)};
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
