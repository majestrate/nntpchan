
-- simple replacements demo

function span_text(class, str)
  return string.format("<div class='%s'>%s</div>", class, str)
end

function wobble_text(str)
  return span_text("wobble", str)
end

function explode_text(str)
  return span_text("explode", str)
end

function psy_text(str)
  return span_text("psy", str)
end

-- generate extra markup
function memeposting(body, prefix)
  body = string.gsub(body, "|(.-)|", wobble_text)
  body = string.gsub(body, "//(.-)\\\\", explode_text)
  body = string.gsub(body, "/@(.-)@\\", psy_text)
  return body
end
