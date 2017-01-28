
-- simple replacements demo

function span_text(str, class)
  return string.format("<span class='%s'>%s</span>", class, str)
end

function wobble_text(str)
  return span_text("wobble", str)
end

-- generate extra markup
function memeposting(prefix, body)
  body = string.gsub(body, "<(%w+)>", wobble_text)
  return body
end
