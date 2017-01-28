
-- simple replacements demo

function span_text(class, str)
  return string.format("<div class='%s'>%s</div>", class, str)
end

function wobble_text(str)
  return span_text("wobble", str)
end

-- generate extra markup
function memeposting(body, prefix)
  body = string.gsub(body, "|(.-)|", wobble_text)
  return body
end
