export type Target = {
  id: string;
  tag: string;
  source: string;
  className:string;
  classKind:string;
  revision:string;
  line: number;
  styles: Record<string, string>;
};
export type Intent = {
  scope: string;
  family: string;
  operation: string;
  direction: string;
  magnitude: string;
  breakpoint: string;
  value?: string;
  confidence: number;
};
export type State = {
 files:string[]|null;executor:string;
  project: string;
  session: string;
  previewURL: string;
  connected: boolean;
  browserConnected: boolean;
  selected: Target | null;
  hasAPIKey: boolean;
  mode: string;
  busy: boolean;
  pending: boolean;
  canUndo: boolean;
  selecting: boolean;
  intents: Intent[] | null;
  diff: string;
  error: string;
  stages: { name: string; ms: number }[] | null;
  changeId: string;
};
const native = () =>
  (
    window as unknown as {
      go?: {
        main?: {
          App?: Record<string, (...args: unknown[]) => Promise<unknown>>;
        };
      };
    }
  ).go?.main?.App;
export const isNative = () => !!native();
export async function call<T = State>(
  action: string,
  value: Record<string, unknown> = {},
): Promise<T> {
  const binding = native();
  if (binding) {
    const mapping: Record<string, [string, unknown[]]> = {
      state: ["GetState", []],
      open: ["OpenProject", [value.path]],
      choose: ["ChooseProject", []],
      start: ["StartPreview", []],
      connect: ["ConnectPreview", [value.url]],
      stage: ["OpenStage", []],
      select: ["SetSelecting", [value.on]],
      offline: ["SetOffline", [value.on]],
      apply: ["ApplyEdit", [value]],
      accept: ["Accept", []],
      undo: ["Undo", []],
      reject: ["Reject", []],
    };
    const [name, args] = mapping[action];
    return (await binding[name](...args)) as T;
  }
  const response = await fetch(
    "/__jev/api/" + action,
    action === "state"
      ? {}
      : {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(value),
        },
  );
  const data = await response.json();
  if (!response.ok) throw Error(data.error || "连接失败");
  return data;
}
