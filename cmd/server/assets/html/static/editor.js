function SetupEditor(editorTabButton, previewTabButton, editorTab, previewTab, content, attachment, cost, limit, submit, suffix) {
  // Markdown Preview
  const parser = new commonmark.Parser();
  const updatePreview = function() {
    previewTab.innerHTML = markdownToHTML(parser, content.value);
  };

  // Tab Handling
  const openEditor = function() {
    editorTab.style.display = "block";
    previewTab.style.display = "none";
    editorTabButton.className = "active";
    previewTabButton.className = "";
  }
  const openPreview = function() {
    updatePreview();
    editorTab.style.display = "none";
    previewTab.style.display = "block";
    editorTabButton.className = "";
    previewTabButton.className = "active";
  }

  // Cost Estimate
  const encoder = new TextEncoder();
  const updateCost = function() {
  var c = encoder.encode(content.value).length;
  if (attachment.files[0]) {
    c = c + attachment.files[0].size;
  }
  cost.innerHTML = c+suffix;
  submit.disabled = c > limit;
  };

  // Attach Handlers
  editorTabButton.onclick = openEditor;
  previewTabButton.onclick = openPreview;
  content.onkeyup = updateCost;
  attachment.onchange = updateCost;

  // Initialization
  openEditor();
  updateCost();
}

function markdownToHTML(parser, markdown) {
  const walker = parser.parse(markdown).walker();
  var event, node;
  var result = "";
  while ((event = walker.next())) {
    node = event.node;
    if (event.entering) {
      switch (node.type) {
        case "document":
          // Do nothing
          break;
        case "paragraph":
          result += '<p class="ucc">';
          break;
        case "text":
          result += node.literal;
          break;
        case "thematic_break":
          result += '<hr class="ucc" />\n';
          break;
        case "softbreak":
          result += ' ';
          break;
        case "linebreak":
          result += '<br />';
          break;
        case "heading":
          result += '<h' + node.level + ' class="ucc">';
          break;
        case "emph":
          result += '<em class="ucc">'
          break;
        case "strong":
          result += '<strong class="ucc">'
          break;
        case "block_quote":
          result += '<blockquote class="ucc">\n';
          break;
        case "code":
          result += '<pre class="ucc"><code class="ucc">';
          break;
        case "code_block":
          result += '<pre class="ucc"><code class="ucc">\n';
          result += node.literal;
          break;
        case "list":
          switch (node.listType) {
            case "ordered":
              if (node.listStart == 1) {
                result += '<ol class="ucc">\n';
              } else {
                result += '<ol class="ucc" start="' + node.listStart + '">\n';
              }
              break;
            case "bullet":
              result += '<ul class="ucc">\n';
              break;
          }
          break;
        case "item":
          result += '<li class="ucc">';
          if (!node.parent.IsTight) {
            result += '\n';
          }
          break;
        case "link":
          result += '<a class="ucc" href="' + node.destination + '" title="' + node.title + '">';
          break;
        default:
          console.log("Entering Unhandled Node:", node.type);
          break;
      }
    } else {
      switch (node.type) {
        case "document":
          // Do nothing
          break;
        case "paragraph":
          result += '</p>\n';
          break;
        case "text":
          // Do nothing
          break;
        case "thematic_break":
          // Do nothing
          break;
        case "heading":
          result += '</h' + node.level + '>\n';
          break;
        case "emph":
          result += '</em>';
          break;
        case "strong":
          result += '</strong>';
          break;
        case "block_quote":
          result += '</blockquote>\n';
          break;
        case "code":
          result += '</code></pre>';
          break;
        case "code_block":
          result += '</code></pre>\n';
          break;
        case "list":
          switch (node.listType) {
            case "ordered":
              result += '</ol>\n';
              break;
            case "bullet":
              result += '</ul>\n';
              break;
          }
          break;
        case "item":
          result += '</li>\n';
          break;
        case "link":
          result += '</a>';
          break;
        default:
          console.log("Exiting Unhandled Node:", node.type);
          break;
      }
    }
  }
  return result;
}
