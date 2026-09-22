import {createRoot} from 'react-dom/client';
import clsx from 'clsx';
import './style.css';
const enabled=true;
function calculateClass(){return 'rounded-lg bg-blue-600 px-4 py-2 text-white'}
function App(){return <main className="mx-auto max-w-5xl px-6 py-10 font-sans text-slate-800">
 <header className="mb-10 flex items-center justify-between border-b border-slate-200 pb-6"><h1 className="text-2xl font-semibold">Jev <span className="text-sm font-normal text-slate-500">Tailwind 工作台</span></h1><span className="text-xs text-slate-500">选择 · 描述 · 接受</span></header>
 <section className="mb-10"><p className="mb-2 text-sm text-slate-500">01 / 精确样式</p><h2 data-jev-id="workspace-title" className="mb-5 text-3xl font-semibold">从一句话，到真实源码。</h2><p className="mb-6 max-w-xl leading-relaxed text-slate-600">选择按钮，将背景换成红色。已有的间距、圆角和交互状态会保留。</p><button data-jev-id="purchase-button" type="button" className="rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 md:px-6 focus:ring-2" onClick={()=>document.querySelector('#notice')!.textContent='已点击'}>立即购买</button><p id="notice" className="mt-2 text-sm"/></section>
 <section className="mb-10"><p className="mb-2 text-sm text-slate-500">02 / 添加内容</p><h2 className="mb-4 text-xl font-semibold">给空白一个开始</h2><div data-jev-id="empty-container" aria-label="可编辑空框" className="h-64 rounded-xl border border-dashed border-slate-300 bg-slate-50"></div></section>
 <section data-jev-id="occupied-container" className="mb-10 rounded-xl border border-slate-200 p-6"><h2 className="text-xl font-semibold">已有内容的容器</h2><p className="mt-3 leading-relaxed text-slate-600">这里保留已有标题和正文。“在中间加按钮”需要先明确布局，不会静默重排内容。</p></section>
 <section data-jev-id="feature-grid" className="mb-10 grid grid-cols-1 gap-4 md:grid-cols-2"><article data-jev-id="feature-one" className="rounded-xl bg-slate-100 p-6"><h3>可靠的作用域</h3><p className="mt-2 text-sm">保留其他状态与断点。</p></article><article className="rounded-xl bg-slate-100 p-6">可撤销的真实源码</article></section>
 <section className="grid gap-3 sm:grid-cols-2"><button data-jev-id="template-button" className={`rounded-lg bg-blue-600 px-4 py-2 text-white`}>静态模板</button><button data-jev-id="clsx-button" className={clsx('rounded-lg px-4 py-2 text-white',enabled && 'bg-blue-600')}>简单 clsx 条件</button><button data-jev-id="conditional-button" className={enabled && 'rounded-lg bg-blue-600 px-4 py-2 text-white'}>简单条件类</button><button data-jev-id="dynamic-button" className={calculateClass()}>动态 class 应拒绝</button></section>
</main>}
createRoot(document.getElementById('root')!).render(<App/>);
