function SetupEditor(form, editorTabButton, previewTabButton, editorTab, previewTab, content, attachment, cost, limit, submit, action, suffix) {
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

  // Markdown Preview
  const parser = new commonmark.Parser();
  const updatePreview = function() {
    previewTab.innerHTML = markdownToHTML(parser, content.value);
  };

  // Tab Switching
  const openPreview = function() {
    updateCost();
    updatePreview();
    editorTab.style.display = "none";
    previewTab.style.display = "block";
    editorTabButton.className = "";
    previewTabButton.className = "active";
    submit.value = action;
  }
  const openEditor = function() {
    editorTab.style.display = "block";
    previewTab.style.display = "none";
    editorTabButton.className = "active";
    previewTabButton.className = "";
    submit.value = "Preview";
  }

  // Submit Button
  const submitPost = function() {
    const action = submit.value;
    switch (action) {
      case "Preview":
        openPreview();
        break;
      case action:
        submit.disabled = true;
        form.submit();
        form.reset();
        break;
      default:
        console.log("Unrecognized Submit Action: " + action);
        break;
    }
  }

  // Attach Handlers
  editorTabButton.onclick = openEditor;
  previewTabButton.onclick = openPreview;
  content.onkeyup = updateCost;
  attachment.onchange = updateCost;
  submit.onclick = submitPost;

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
          // Do Nothing
          break;
        case "paragraph":
          const grandparent = node.parent.parent;
          if (grandparent !== null && grandparent.type === "list") {
            if (grandparent.listTight) {
              break;
            }
          }
          result += '<p class="ucc">';
          break;
        case "text":
          result += escape(node.literal);
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
          result += '<code class="ucc">';
          result += escape(node.literal);
          result += '</code>';
          break;
        case "code_block":
          result += '<pre class="ucc"><code class="ucc">\n';
          result += escape(node.literal);
          result += '</code></pre>\n';
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
          if (!node.parent.listTight) {
            result += '\n';
          }
          break;
        case "link":
          result += '<a class="ucc" href="' + node.destination + '"';
          if (node.title) {
            result += ' title="' + escape(node.title) + '"';
          }
          result += '>';
          break;
        case "html_inline":
          // Not Supported
          break;
        default:
          console.log("Entering Unhandled Node: " + node.type);
          break;
      }
    } else {
      switch (node.type) {
        case "document":
          // Do Nothing
          break;
        case "paragraph":
          const grandparent = node.parent.parent;
          if (grandparent !== null && grandparent.type === "list") {
            if (grandparent.listTight) {
              break;
            }
          }
          result += '</p>\n';
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
        case "html_inline":
          // Not Supported
          break;
        default:
          console.log("Exiting Unhandled Node: " + node.type);
          break;
      }
    }
  }
  return result;
}

function escape(unsafe) {
  return unsafe
     .replace(/&/g, "&amp;")
     .replace(/</g, "&lt;")
     .replace(/>/g, "&gt;")
     .replace(/"/g, "&#34;")
     .replace(/'/g, "&#39;");
 }
