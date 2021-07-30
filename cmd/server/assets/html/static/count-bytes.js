function AttachByteCount(content, attachment, destination, limit, submit, suffix) {
  if (typeof TextEncoder === "undefined") {
    console.log("TextEncoder Not Defined");
  } else {
    const encoder = new TextEncoder();
    var trigger = function() {
      var cost = encoder.encode(content.value).length;
      if (attachment.files[0]) {
        cost = cost + attachment.files[0].size;
      }
      destination.innerHTML = cost+suffix;
      submit.disabled = cost > limit;
    };
    content.onkeyup = trigger;
    attachment.onchange = trigger;
    trigger();
  }
}