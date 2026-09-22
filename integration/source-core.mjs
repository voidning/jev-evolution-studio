import MagicString from 'magic-string';
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import * as babel from '@babel/parser';
import * as recast from 'recast';
import {twMerge} from 'tailwind-merge';
export const sha=s=>crypto.createHash('sha256').update(s).digest('hex');
const parser={parse:source=>babel.parse(source,{sourceType:'module',plugins:['jsx','typescript'],tokens:true})};
const fail=message=>{throw Error('不支持：'+message)};
const idOf=n=>n?.openingElement?.attributes.find(a=>a.type==='JSXAttribute'&&a.name.name==='data-jev-id')?.value?.value;
const nameOf=n=>n.type==='JSXIdentifier'?n.name:n.type==='JSXMemberExpression'?nameOf(n.object)+'.'+nameOf(n.property):'';
function walk(node,visit,ancestors=[]){if(!node||typeof node!=='object')return;if(node.type)visit(node,ancestors);for(const [key,value] of Object.entries(node)){if(['loc','tokens','comments','original'].includes(key))continue;if(Array.isArray(value))value.forEach(v=>walk(v,visit,[...ancestors,node]));else if(value&&typeof value==='object')walk(value,visit,[...ancestors,node]);}}
export function contained(root,relative){if(path.isAbsolute(relative))fail('源路径必须位于项目内');const file=path.resolve(root,relative);if(!file.startsWith(root+path.sep))fail('源文件越界');const real=fs.realpathSync(file);if(real!==file)fail('不支持符号链接源文件');return real;}
function classLeaves(element,ast){
 const attr=element.openingElement.attributes.find(a=>a.type==='JSXAttribute'&&a.name.name==='className');
 if(element.openingElement.attributes.some(a=>a.type==='JSXAttribute'&&a.name.name==='style'))fail('内联 style 可能覆盖 Tailwind；请先改为静态 className');
 if(element.openingElement.attributes.some(a=>a.type==='JSXSpreadAttribute'))fail('JSX spread 可能覆盖 className，不能可靠定位');
 if(!attr)return {attr:null,leaves:[],kind:'missing'};
 const leaves=[];let kind='literal';
 const imports=new Set();for(const n of ast.program.body){if(n.type==='ImportDeclaration'&&['clsx','classnames'].includes(n.source.value))for(const s of n.specifiers)imports.add(s.local.name)}
 function value(n,conditional=false){
  if(n?.type==='StringLiteral'){leaves.push({node:n,text:n.value,conditional});return;}
  if(n?.type==='TemplateLiteral'&&n.expressions.length===0){leaves.push({node:n,text:n.quasis[0].value.cooked,conditional});kind='template';return;}
  if(n?.type==='LogicalExpression'&&n.operator==='&&'){if(!['StringLiteral','TemplateLiteral'].includes(n.right.type))fail('仅支持 condition && 静态字符串');kind='conditional';return value(n.right,true)}
  if(n?.type==='CallExpression'&&n.callee.type==='Identifier'&&imports.has(n.callee.name)){kind='clsx';for(const arg of n.arguments){if(!['StringLiteral','TemplateLiteral','LogicalExpression'].includes(arg.type))fail('clsx/classnames 仅支持静态字符串与简单条件类');value(arg)}return;}
  fail('动态 className 无法可靠修改；请改用静态字符串、无插值模板或简单 clsx 条件字符串');
 }
 value(attr.value?.type==='JSXExpressionContainer'?attr.value.expression:attr.value);
 return {attr,leaves,kind};
}
function readFiles(root){const files=[];function descend(dir){for(const d of fs.readdirSync(dir,{withFileTypes:true})){if(d.isSymbolicLink())continue;const f=path.join(dir,d.name);if(d.isDirectory())descend(f);else if(/\.(jsx|tsx|css)$/.test(f))files.push(f)}}descend(path.join(root,'src'));return files;}
export function inspectProject(root){
 root=fs.realpathSync(root);const pkg=JSON.parse(fs.readFileSync(path.join(root,'package.json'),'utf8'));const deps={...pkg.dependencies,...pkg.devDependencies};const tailwind=!!deps.tailwindcss;
 const targets=[],tokens={},components=[{id:'button',component:'button',defaultContent:'按钮',allowedParents:['container'],source:'',importFrom:'',variants:[]}];const seen=new Set();
 for(const file of readFiles(root)){
  const source=fs.readFileSync(file,'utf8');
  if(file.endsWith('.css')){for(const theme of source.matchAll(/@theme[^{}]*\{([^}]+)\}/g)){for(const m of theme[1].matchAll(/--color-([\w-]+)\s*:\s*([^;]+);/g))tokens[m[1]]=m[2].trim()}continue;}
  const ast=parser.parse(source);const relative=path.relative(root,file).split(path.sep).join('/');
  let hasButton=false,exported=false;
  walk(ast,n=>{if(n.type==='JSXElement'&&nameOf(n.openingElement.name)==='button')hasButton=true;if(n.type==='ExportNamedDeclaration'&&(n.declaration?.id?.name==='Button'||n.declaration?.declarations?.some(d=>d.id.name==='Button')||n.specifiers?.some(s=>s.exported.name==='Button')))exported=true});
  if(hasButton&&exported&&relative.startsWith('src/components/'))components.push({id:'project-button',component:'Button',source:relative,importFrom:relative,defaultContent:'按钮',defaultProps:{type:'button'},allowedParents:['container'],variants:['default']});
  walk(ast,(n,ancestors)=>{if(n.type!=='JSXElement')return;const id=idOf(n);if(id===undefined)return;if(typeof id!=='string'||!/^[\w-]{1,100}$/.test(id)||seen.has(id))fail('ID 必须是唯一静态字符串');seen.add(id);
   const tag=nameOf(n.openingElement.name);if(!/^[a-z]/.test(tag))fail('只支持可定位的原生元素；不直接修改第三方组件内部样式');
   let info;try{info=classLeaves(n,ast)}catch(e){info={kind:'dynamic',leaves:[]}}
   targets.push({id,tag,source:relative,line:n.loc.start.line,start:n.start,end:n.end,revision:sha(source),fingerprint:sha(source.slice(n.start,n.end)),classKind:info.kind,className:info.leaves.map(l=>l.text).join(' '),parent:[...ancestors].reverse().filter(a=>a.type==='JSXElement').map(idOf).find(Boolean)||'',repeated:ancestors.some(a=>a.type==='CallExpression'&&a.callee.type==='MemberExpression'&&a.callee.property.name==='map'),empty:n.children.every(c=>c.type==='JSXText'&&!c.value.trim()||c.type==='JSXExpressionContainer'&&c.expression.type==='JSXEmptyExpression')});
  });
 }
 for(const name of ['tailwind.config.js','tailwind.config.cjs','tailwind.config.ts']){const file=path.join(root,name);if(!fs.existsSync(file))continue;try{walk(parser.parse(fs.readFileSync(file,'utf8')),n=>{if(n.type==='ObjectProperty'&&(n.key.name||n.key.value)==='colors'&&n.value.type==='ObjectExpression'){for(const p of n.value.properties){if(p.type==='ObjectProperty'&&p.value.type==='StringLiteral')tokens[p.key.name||p.key.value]=p.value.value}}})}catch{}}
 return {root,version:1,tailwind,targets,tokens,components};
}
const splitToken=token=>{let depth=0,at=-1;for(let i=0;i<token.length;i++){if(token[i]==='['||token[i]==='(')depth++;if(token[i]===']'||token[i]===')')depth--;if(token[i]===':'&&depth===0)at=i;}return {prefix:token.slice(0,at+1),base:token.slice(at+1)}};
const textSizes=['xs','sm','base','lg','xl','2xl','3xl','4xl','5xl','6xl','7xl','8xl','9xl'];
function group(c){
 if(c.startsWith('bg-')&&twMerge(c,'bg-red-600')==='bg-red-600')return 'background';
 if(/^text-(xs|sm|base|lg|xl|[2-9]xl)(\/.*)?$/.test(c)||/^text-\[(length:)?[\d.]+(px|rem|em)\]$/.test(c))return 'fontSize';
 if(/^text-(left|center|right|justify|start|end)$/.test(c))return 'textAlign';if(c.startsWith('text-')&&twMerge(c,'text-red-600')==='text-red-600')return 'color';
 if(/^font-(thin|extralight|light|normal|medium|semibold|bold|extrabold|black)$/.test(c))return 'fontWeight';
 if(/^border(-(0|2|4|8))?$/.test(c))return 'border';if(/^border-(solid|dashed|dotted|double|none)$/.test(c))return 'border';if(c.startsWith('border-')&&!/^border-[trblxy]-/.test(c))return 'borderColor';
 if(/^p[trblxyse]?-/.test(c))return 'padding';if(/^-?m[trblxyse]?-/.test(c))return 'margin';if(/^gap(-[xy])?-/.test(c))return 'gap';
 if(/^max-w-/.test(c))return 'maxWidth';if(/^w-/.test(c))return 'width';if(/^rounded/.test(c))return 'radius';if(/^shadow/.test(c))return 'shadow';if(/^opacity-/.test(c))return 'opacity';if(/^grid-cols-/.test(c))return 'columns';if(/^items-/.test(c))return 'items';if(/^justify-/.test(c))return 'justify';if(['hidden','block','inline-block','inline','flex','inline-flex','grid','inline-grid'].includes(c))return 'display';if(/^flex-(row|col)/.test(c))return 'axis';if(/^place-items-/.test(c))return 'placeItems';return '';
}
function prefix(in_){const bp={all:'',mobile:'max-md:',desktop:'md:',sm:'sm:',md:'md:',lg:'lg:'}[in_.breakpoint];if(bp===undefined)fail('未知断点');if(in_.state&&!['hover','focus'].includes(in_.state))fail('未注册状态');return bp+(in_.state?in_.state+':':'');}
function colorClass(c,channel,meta){const base={background:'bg',text:'text',border:'border'}[channel];if(!base)fail('颜色通道不支持');if(c.kind==='token'){if(!(c.value in meta.tokens))fail('项目未定义设计令牌 '+c.value);return base+'-'+c.value}
 if(c.kind==='literal'&&/^(red|blue|green|yellow|orange|purple|pink|slate|gray|black|white|transparent)$/.test(c.value))return base+'-'+c.value+(['black','white','transparent'].includes(c.value)?'':'-'+(c.shade||'600'));
 if(c.kind==='literal'&&/^(#[\da-fA-F]{3,8}|(?:rgb|hsl|oklch)\([\d\s.,%/+-]+\))$/.test(c.value)){return base+'-['+c.value.trim().replace(/\s+/g,'_')+']'}fail('颜色不是允许的字面值或项目令牌');}
function editsFor(in_,classes,meta){
 const current=g=>classes.map(splitToken).filter(c=>c.prefix===prefix(in_)&&group(c.base)===g).map(c=>c.base);
 const relative=(g,base,def)=>{const old=current(g);const adjust=v=>{const num=Number(v.split('-').at(-1));if(!Number.isFinite(num))fail('当前尺寸不是可修改的数值刻度');return Math.max(0,Math.min(96,num+(in_.direction==='decrease'?-1:1)))};return old.length?old.map(v=>v.slice(0,v.lastIndexOf('-')+1)+adjust(v)):[base+'-'+adjust(base+'-'+def)]};
 const op=in_.operation;let changes=[];
 if(op==='set_color'){changes=[[{background:'background',text:'color',border:'borderColor'}[in_.arguments.channel],colorClass(in_.arguments.color,in_.arguments.channel,meta)]];}
 else if(['density','padding','margin','gap'].includes(op)){const groups=op==='density'?['padding','gap']:[op];changes=groups.map(g=>[g,relative(g,{padding:'p',margin:'m',gap:'gap'}[g],4).join(' ')]);}
 else if(op==='fontSize'){const c=current('fontSize')[0]||'text-base';const at=textSizes.indexOf(c.slice(5));if(at<0)fail('自定义字号需要明确数值，不能猜测');changes=[['fontSize','text-'+textSizes[Math.max(0,Math.min(textSizes.length-1,at+(in_.direction==='decrease'?-1:1)))]]];}
 else if(op==='fontWeight'){const values=['thin','extralight','light','normal','medium','semibold','bold','extrabold','black'];const at=values.indexOf((current('fontWeight')[0]||'font-normal').slice(5));changes=[['fontWeight','font-'+values[Math.max(0,Math.min(8,at+(in_.direction==='decrease'?-1:1)))]]];}
 else if(op==='textAlign')changes=[['textAlign','text-'+in_.value]];
 else if(op==='columns')changes=[['display','grid'],['columns','grid-cols-'+in_.value]];
 else if(op==='axis')changes=[['display','flex'],['axis','flex-'+(in_.value==='column'?'col':'row')]];
 else if(op==='center')changes=[['display','flex'],['items','items-center'],['justify','justify-center']];
 else if(op==='width')changes=[['width','w-full'],['maxWidth',{'narrow':'max-w-xl','normal':'max-w-5xl','wide':'max-w-7xl','full':'max-w-full'}[in_.value]]];
 else if(op==='radius')changes=[['radius',{'sharp':'rounded-sm','medium':'rounded-lg','round':'rounded-3xl'}[in_.value]]];
 else if(op==='border')changes=[['border',in_.value==='show'?'border':'border-0']];
 else if(op==='visibility')changes=[['display',in_.value==='hidden'?'hidden':(in_.value==='visible'?'block':in_.value)]];
 else if(op==='set_utility')changes=[[in_.arguments.channel,in_.arguments.utility]];
 else if(op==='background')changes=[['background',{'stronger':'bg-slate-700','softer':'bg-slate-100','transparent':'bg-transparent'}[in_.value]]];
 else if(op==='emphasis')changes=[['background',in_.value==='primary'?(meta.tokens.primary?'bg-primary':'bg-blue-600'):in_.value==='muted'?'bg-slate-100':'bg-white'],['color',in_.value==='primary'?'text-white':'text-slate-700'],['fontWeight',in_.value==='primary'?'font-bold':'font-normal']];
 else fail('Tailwind 执行器不支持此操作');
 if(changes.some(([g,c])=>!g||!c||c.includes('undefined')))fail('操作参数无效');return changes;
}
function changeClasses(element,ast,in_,meta){const info=classLeaves(element,ast);if(info.leaves.length===0){const literal=recast.types.builders.stringLiteral('');const attr=recast.types.builders.jsxAttribute(recast.types.builders.jsxIdentifier('className'),literal);element.openingElement.attributes.push(attr);info.leaves.push({node:literal,text:'',conditional:false})}
 const all=info.leaves.flatMap(l=>l.text.trim().split(/\s+/).filter(Boolean));const pre=prefix(in_);const changes=editsFor(in_,all,meta);
 for(const [channel,value] of changes){const matches=info.leaves.filter(l=>l.text.split(/\s+/).some(token=>{const c=splitToken(token);return c.prefix===pre&&group(c.base)===channel}));if(matches.length>1&&matches.some(l=>l.conditional))fail('同一属性分布在多个条件分支，需明确编辑分支');const dest=matches[0]||info.leaves.find(l=>!l.conditional)||info.leaves[0];
  for(const leaf of matches){leaf.text=leaf.text.split(/\s+/).filter(token=>{const c=splitToken(token);return !(c.prefix===pre&&group(c.base)===channel)}).join(' ')}
  // Only the requested group+variant is merged; all other variants remain byte-for-byte tokens.
  dest.text=(dest.text+' '+twMerge(value.split(/\s+/).map(v=>pre+v).join(' '))).trim();
 }
 for(const leaf of info.leaves){if(leaf.node.type==='StringLiteral'){leaf.node.value=leaf.text;if(leaf.node.extra)delete leaf.node.extra;}else{leaf.node.quasis[0].value={raw:leaf.text.replace(/`/g,'\\`'),cooked:leaf.text};}}
 return info.leaves.map(l=>l.text).join(' ');
}
function insertButton(element,ast,in_,meta){const tag=nameOf(element.openingElement.name);if(!['div','section','main','article','aside','header','footer'].includes(tag))fail('所选元素不是允许插入按钮的容器');if(!element.children.every(c=>c.type==='JSXText'&&!c.value.trim()||c.type==='JSXExpressionContainer'&&c.expression.type==='JSXEmptyExpression'))fail('歧义：容器已有内容。请明确“作为现有内容的一部分居中”还是“悬浮在容器正中央”；本版仅支持空容器，可选择空框后重试。');
 const info=classLeaves(element,ast);if(info.kind==='dynamic'||info.leaves.some(l=>l.conditional))fail('插入要求静态容器 className');
 for(const token of info.leaves.flatMap(l=>l.text.split(/\s+/))){const c=splitToken(token);if(c.prefix&&['display','columns','items','justify','placeItems','axis'].includes(group(c.base)))fail('容器包含响应式或状态布局，不能保证所有视口居中；请先简化容器布局');}
 for(const [channel,utility] of [['display','grid'],['columns','grid-cols-1'],['items','items-center'],['justify','justify-center'],['placeItems','place-items-center']])changeClasses(element,ast,{...in_,operation:'set_utility',arguments:{channel,utility},breakpoint:'all'},meta);
 const b=recast.types.builders;const candidates=meta.components.filter(c=>c.component==='Button'&&c.source!==in_.source);if(candidates.length>1)fail('存在多个 Button 组件，无法唯一选择注册组件');const registered=candidates[0];let name='button',attributes=[];
 if(registered){const absolute=contained(meta.root,registered.source);let from=path.relative(path.dirname(path.join(meta.root,in_.source)),absolute).split(path.sep).join('/').replace(/\.[jt]sx$/,'');if(!from.startsWith('.'))from='./'+from;
  let local;for(const node of ast.program.body){if(node.type==='ImportDeclaration'&&(node.source.value===from||node.source.value===registered.importFrom)){const spec=node.specifiers.find(s=>s.type==='ImportSpecifier'&&s.imported.name==='Button');if(spec)local=spec.local.name}}
  if(!local){local='JevButton';const used=new Set();walk(ast,n=>{if(n.type==='Identifier')used.add(n.name)});for(let i=2;used.has(local);i++)local='JevButton'+i;ast.program.body.unshift(b.importDeclaration([b.importSpecifier(b.identifier('Button'),b.identifier(local))],b.stringLiteral(from)))}name=local;
 }else{attributes.push(b.jsxAttribute(b.jsxIdentifier('className'),b.stringLiteral('rounded-lg '+(meta.tokens.primary?'bg-primary ':'bg-blue-600 ')+'px-4 py-2 '+(meta.tokens['primary-foreground']?'text-primary-foreground':'text-white'))));}
 attributes.unshift(b.jsxAttribute(b.jsxIdentifier('type'),b.stringLiteral('button')));
 const label=in_.arguments.label||'按钮';if(typeof label!=='string'||label.length>120)fail('按钮文案长度必须不超过120');
 // String literal child avoids interpreting <, &, braces or quotes as JSX syntax.
 const child=b.jsxElement(b.jsxOpeningElement(b.jsxIdentifier(name),attributes,false),b.jsxClosingElement(b.jsxIdentifier(name)),[b.jsxExpressionContainer(b.stringLiteral(label))]);
 element.openingElement.selfClosing=false;element.closingElement=b.jsxClosingElement(element.openingElement.name);element.children.push(b.jsxText('\n  '),child,b.jsxText('\n'));
 return {label,component:name};
}
export function createSourcePatch(request){
 const {root,targets,intents}=request,meta=inspectProject(root);if(!meta.tailwind)fail('该项目未使用 Tailwind');const files=[],checks=[];
 for(const relative of [...new Set(targets.map(t=>t.source))]){
  const file=contained(root,relative),before=fs.readFileSync(file,'utf8'),ast=recast.parse(before,{parser});
  const text=new MagicString(before),originalBody=new Set(ast.program.body);
  for(const target of targets.filter(t=>t.source===relative)){
   const fresh=meta.targets.find(t=>t.id===target.id);if(!fresh||fresh.revision!==target.revision||fresh.fingerprint!==target.fingerprint)fail('源码 revision 或节点指纹已过期，请重新选择');if(fresh.repeated)fail('目标来自 map()，修改会影响共享实例；本版拒绝单实例编辑');
   let element;recast.types.visit(ast,{visitJSXElement(p){if(idOf(p.node)===target.id)element=p.node;this.traverse(p)}});if(!element)fail('目标节点已消失');
   const originalClass=element.openingElement.attributes.find(a=>a.type==='JSXAttribute'&&a.name.name==='className');
   for(const in_ of intents){if(in_.operation==='insert_child'){const child=insertButton(element,ast,{...in_,source:relative},meta);checks.push({id:target.id,...child});}else{const className=changeClasses(element,ast,in_,meta);checks.push({id:target.id,className,classKind:fresh.classKind});}}
   const replacement=intents.some(i=>i.operation==='insert_child')||!originalClass?element:originalClass;
   text.overwrite(replacement.start,replacement.end,recast.print(replacement,{quote:'double',tabWidth:2}).code);
  }
  const addedImports=ast.program.body.filter(n=>!originalBody.has(n));
  if(addedImports.length){const existing=[...originalBody].filter(n=>n.type==='ImportDeclaration');const at=existing.at(-1)?.end||ast.program.directives?.at(-1)?.end||0;text.appendLeft(at,'\n'+addedImports.map(n=>recast.print(n,{quote:'double'}).code).join('\n')+'\n');}
  const after=text.toString();parser.parse(after);if(before===after)fail('没有需要修改的源码');files.push({path:relative,before,after});
 }
 const finalChecks=[...checks.reduce((m,c)=>m.set(c.id,{...m.get(c.id),...c}),new Map()).values()];
 return {files,checks:finalChecks,executor:intents.some(i=>i.operation==='insert_child')?'jsx':'tailwind'};
}
