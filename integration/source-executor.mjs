import {createSourcePatch,inspectProject} from './source-core.mjs';
async function main(){
let input='';for await (const chunk of process.stdin)input+=chunk;
try{const request=JSON.parse(input);const result=request.action==='inspect'?inspectProject(request.root):createSourcePatch(request);process.stdout.write(JSON.stringify(result));}
catch(error){process.stdout.write(JSON.stringify({error:error.message}));process.exitCode=1;}

}
void main();
