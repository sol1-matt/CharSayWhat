//go:build windows

package main

import webview2 "github.com/jchv/go-webview2"

const htmlPage = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<style>
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
html,body{height:100%;overflow:hidden}
body{
  display:flex;flex-direction:column;
  background:#1e1e1e;color:#d4d4d4;
  font-family:'Consolas','Cascadia Code','Courier New',monospace;
}
#input{
  flex:0 0 40%;
  background:#252526;color:#d4d4d4;
  border:none;border-bottom:2px solid #007acc;
  padding:12px 16px;font-size:14px;font-family:inherit;
  resize:none;outline:none;line-height:1.5;
}
#input::placeholder{color:#5a5a5a}
#output{
  flex:1;
  overflow:auto;
  padding:12px 16px;
}

table{
  border-collapse:collapse;
  table-layout:fixed;
}
th.label{
  width:52px;
  text-align:right;
  padding-right:10px;
  font-size:11px;
  font-weight:normal;
  color:#666;
  white-space:nowrap;
  vertical-align:middle;
}
td{
  width:56px;
  border:1px solid #2d2d2d;
  text-align:center;
  padding:4px 2px;
}

tr.row-pos td{
  font-size:11px;
  color:#666;
  vertical-align:middle;
}
tr.row-char td{
  font-size:22px;
  color:#d4d4d4;
  vertical-align:middle;
  padding:6px 2px;
}
tr.row-type td{
  font-size:11px;
  vertical-align:top;
  word-break:break-word;
  white-space:normal;
  line-height:1.4;
}

.lowercase  {color:#9cdcfe}
.uppercase  {color:#c586c0}
.digit      {color:#b5cea8}
.space      {color:#484848}
.newline    {color:#484848}
.tab        {color:#484848}
.punctuation{color:#ce9178}
.symbol     {color:#4fc1ff}
.other      {color:#d4d4d4}
</style>
</head>
<body>
<textarea id="input" placeholder="Paste or type text here..."></textarea>
<div id="output"></div>
<script>
const RE_LOWER  = /\p{Ll}/u;
const RE_UPPER  = /\p{Lu}/u;
const RE_DIGIT  = /\p{Nd}/u;
const RE_PUNCT  = /\p{P}/u;
const RE_SYMBOL = /\p{S}/u;

function classify(ch) {
  if (ch === ' ')          return 'space';
  if (ch === '\n')         return 'newline';
  if (ch === '\t')         return 'tab';
  if (RE_LOWER.test(ch))  return 'lowercase';
  if (RE_UPPER.test(ch))  return 'uppercase';
  if (RE_DIGIT.test(ch))  return 'digit';
  if (RE_PUNCT.test(ch))  return 'punctuation';
  if (RE_SYMBOL.test(ch)) return 'symbol';
  return 'other';
}

function display(ch) {
  if (ch === ' ')  return ' ';
  if (ch === '\n') return '\\n';
  if (ch === '\t') return '\\t';
  return ch;
}

function esc(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function analyze() {
  const text = document.getElementById('input').value;
  const out  = document.getElementById('output');
  if (!text) { out.innerHTML = ''; return; }

  const chars = [...text].map((ch, i) => ({
    pos:  String(i + 1),
    disp: display(ch),
    type: classify(ch),
  }));

  let posRow  = '<tr class="row-pos"><th class="label">Pos</th>';
  let charRow = '<tr class="row-char"><th class="label">Char</th>';
  let typeRow = '<tr class="row-type"><th class="label">Type</th>';

  for (const c of chars) {
    posRow  += '<td>' + esc(c.pos)  + '</td>';
    charRow += '<td>' + esc(c.disp) + '</td>';
    typeRow += '<td class="' + c.type + '">' + c.type + '</td>';
  }

  out.innerHTML =
    '<table>' +
    posRow  + '</tr>' +
    charRow + '</tr>' +
    typeRow + '</tr>' +
    '</table>';
}

document.getElementById('input').addEventListener('input', analyze);
</script>
</body>
</html>`

func runGUI() {
	w := webview2.New(false)
	if w == nil {
		return
	}
	defer w.Destroy()
	w.SetTitle("CharSayWhat")
	w.SetSize(900, 600, webview2.HintNone)
	w.SetHtml(htmlPage)
	setWindowIconFromExe(uintptr(w.Window()))
	w.Run()
}
