
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

function flashy_text(str)
  return string.format("<span class='flashy'>%s</span>", str)
end

function psy_text(str)
  return string.format("<span class='psy'>%s</span>", str)
end

function red_text(str)
  return string.format("<span class='redtext'>%s</span>", str)
end

function pre_text(str)

  str = str:gsub("%(", "&#40;")
  str = str:gsub("%)", "&#41;")
  str = str:gsub("%[", "&#91;")
  str = str:gsub("\\", "&#92;")
  str = str:gsub("//", "&#47;&#47;")
  str = str:gsub("%]", "&#93;")
  str = str:gsub("==", "&#61;&#61;")
  str = str:gsub("@", "&#64;")
  return span_text("code", str:gsub("%|", "&#124;"))
end

-- generate extra markup
function memeposting(body, prefix)
  body = string.gsub(body, "`(.-)`", pre_text)
  body = string.gsub(body, "%(%(%((.-)%)%)%)", function(str) return string.format("<div class='nazi' style='background-image: url(%sstatic/nazi.png);'>%s</div>", prefix, str) end)
  body = string.gsub(body, "|(.-)|", wobble_text)
  body = string.gsub(body, "%[%[%[(.-)%]%]%]", flashy_text)
  body = string.gsub(body, "==(.-)==", red_text)
  body = string.gsub(body, "@@(.-)@@", psy_text)
  body = string.gsub(body, "//(.-)\\\\", explode_text)
  body = string.gsub(body, "/@(.-)@\\", psy_text)
  return body
end
