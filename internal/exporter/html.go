package exporter

import "html/template"

var htmlTemplate = template.Must(template.New("html").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
:root { color-scheme: light dark; --bg:#f8f8f5; --fg:#202124; --muted:#6b6f76; --line:#d8d9d3; --panel:#ffffff; --accent:#0b6f6a; --err:#b3261e; --code:#f0f1ec; }
@media (prefers-color-scheme: dark) { :root { --bg:#171817; --fg:#eceee8; --muted:#aeb4ad; --line:#343832; --panel:#20221f; --accent:#53c2b8; --err:#ff8a80; --code:#151714; } }
* { box-sizing: border-box; }
body { margin:0; background:var(--bg); color:var(--fg); font:14px/1.45 ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
header { position:sticky; top:0; z-index:2; padding:18px 24px 14px; background:color-mix(in srgb, var(--bg) 92%, transparent); border-bottom:1px solid var(--line); backdrop-filter:blur(10px); }
h1 { margin:0 0 10px; font-size:22px; letter-spacing:0; }
.meta { display:flex; flex-wrap:wrap; gap:10px 18px; color:var(--muted); }
.meta b, .pill b { color:var(--fg); font-weight:600; }
.toolbar { margin-top:14px; display:flex; gap:10px; align-items:center; }
input { width:min(680px, 100%); padding:9px 11px; border:1px solid var(--line); border-radius:6px; background:var(--panel); color:var(--fg); font:inherit; }
main { padding:18px 24px 40px; max-width:1180px; margin:0 auto; }
.section-title { margin:22px 0 10px; color:var(--muted); font-size:12px; font-weight:700; letter-spacing:.08em; text-transform:uppercase; }
.grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(160px,1fr)); gap:10px; }
.pill { padding:10px 12px; border:1px solid var(--line); border-radius:6px; background:var(--panel); }
.event { border:1px solid var(--line); border-radius:6px; background:var(--panel); margin:10px 0; overflow:hidden; }
.event[hidden] { display:none; }
.head { display:grid; grid-template-columns:72px 110px minmax(120px,1fr) 105px 100px; gap:10px; align-items:center; padding:10px 12px; border-bottom:1px solid var(--line); }
.seq { color:var(--muted); font-variant-numeric:tabular-nums; }
.dir { font-family:ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; color:var(--accent); }
.method { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.status { text-transform:uppercase; font-size:12px; font-weight:700; color:var(--muted); }
.status.error { color:var(--err); }
.time { color:var(--muted); font-variant-numeric:tabular-nums; }
details { padding:10px 12px; }
summary { cursor:pointer; color:var(--accent); user-select:none; }
pre { margin:10px 0 0; padding:12px; overflow:auto; border-radius:6px; background:var(--code); font:12px/1.45 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.empty { color:var(--muted); padding:30px 0; }
@media (max-width:720px) { header, main { padding-left:14px; padding-right:14px; } .head { grid-template-columns:56px 70px 1fr; } .status, .time { display:none; } }
</style>
</head>
<body>
<header>
  <h1 id="title">mcpsnoop export</h1>
  <div class="meta" id="meta"></div>
  <div class="toolbar"><input id="q" type="search" placeholder="Search method, id, tool, status, direction, or JSON"></div>
</header>
<main>
  <div class="section-title">Summary</div>
  <div class="grid" id="summary"></div>
  <div class="section-title">Events</div>
  <div id="events"></div>
</main>
<script>
const data = {{.Data}};
const fmtTime = (s) => s ? new Date(s).toLocaleString() : "";
const textOf = (v) => v == null ? "" : (typeof v === "string" ? v : JSON.stringify(v, null, 2));
const compact = (v) => v == null ? "" : (typeof v === "string" ? v : JSON.stringify(v));
const esc = (s) => String(s ?? "").replace(/[&<>"']/g, ch => ({ "&":"&amp;", "<":"&lt;", ">":"&gt;", '"':"&quot;", "'":"&#39;" }[ch]));
document.title = "mcpsnoop " + (data.session.label || data.session.id);
document.getElementById("title").textContent = data.session.label || data.session.id;
document.getElementById("meta").innerHTML = [
  "<span><b>" + esc(data.session.id) + "</b></span>",
  "<span>" + esc(fmtTime(data.session.first)) + "</span>",
  "<span>" + data.events.length + " frames</span>",
  "<span>" + data.calls.length + " calls</span>"
].join("");
document.getElementById("summary").innerHTML = [
  ["Requests", data.session.requests],
  ["Responses", data.session.responses],
  ["Notifications", data.session.notifications],
  ["Errors", data.session.errors],
  ["Pending", data.session.pending],
  ["Protocol", data.capabilities?.protocol_version || ""]
].map(([k,v]) => "<div class=\"pill\">" + esc(k) + "<br><b>" + esc(v) + "</b></div>").join("");
const calls = data.calls || [];
const events = data.events || [];
const eventSearch = (ev) => {
  const call = ev.call_index == null ? null : calls[ev.call_index];
  return [
    ev.seq, ev.direction, ev.kind, ev.method, ev.id, ev.text, compact(ev.raw),
    call?.method, call?.status, call?.tool_name, call?.id, compact(call?.params), compact(call?.result), compact(call?.error)
  ].join(" ").toLowerCase();
};
const renderEvent = (ev) => {
  const call = ev.call_index == null ? null : calls[ev.call_index];
  const status = call?.status || "";
  const raw = ev.text || textOf(ev.raw);
  const callBlock = call ? "<details><summary>Correlated call</summary><pre>" + esc(JSON.stringify(call, null, 2)) + "</pre></details>" : "";
  return "<article class=\"event\" data-search=\"" + esc(eventSearch(ev)) + "\">" +
    "<div class=\"head\">" +
      "<div class=\"seq\">#" + ev.seq + "</div>" +
      "<div class=\"dir\">" + esc(ev.direction) + " " + esc(ev.kind) + "</div>" +
      "<div class=\"method\">" + esc(ev.method || ev.id || ev.text || "") + "</div>" +
      "<div class=\"status " + (status === "error" ? "error" : "") + "\">" + esc(status) + "</div>" +
      "<div class=\"time\">" + esc(fmtTime(ev.timestamp)) + "</div>" +
    "</div>" +
    "<details open><summary>Frame</summary><pre>" + esc(raw) + "</pre></details>" +
    callBlock +
  "</article>";
};
const list = document.getElementById("events");
list.innerHTML = events.length ? events.map(renderEvent).join("") : "<div class=\"empty\">No events</div>";
document.getElementById("q").addEventListener("input", (e) => {
  const q = e.target.value.trim().toLowerCase();
  for (const row of document.querySelectorAll(".event")) {
    row.hidden = q && !row.dataset.search.includes(q);
  }
});
</script>
</body>
</html>
`))
