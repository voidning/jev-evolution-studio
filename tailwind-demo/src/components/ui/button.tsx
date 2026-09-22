import type {ButtonHTMLAttributes} from 'react';
export function Button({className='',...props}:ButtonHTMLAttributes<HTMLButtonElement>){return <button className={'rounded-lg bg-primary px-4 py-2 text-primary-foreground '+className} {...props}/>}
