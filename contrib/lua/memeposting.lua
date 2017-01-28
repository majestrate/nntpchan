
-- simple replacements demo

function span_text(class, str)
  return string.format("<div class='%s'>%s</div>", class, str)
end

function wobble_text(str)
  return span_text("wobble", str)
end

-- generate extra markup
function memeposting(body, prefix)
  local nums = 1
  for nums > 0 do
    body, nums = string.gsub(body, "|(.*)|", wobble_text, 1)
  end
  return body
end
