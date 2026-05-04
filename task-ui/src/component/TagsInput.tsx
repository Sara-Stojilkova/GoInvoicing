import { useState } from "react";

interface TagsInputProps {
  value: string[];
  onChange: (tags: string[]) => void;
  disabled?: boolean;
}

function tagColorClass(tag: string): string {
  let hash = 0;
  for (let i = 0; i < tag.length; i++) {
    hash = (hash * 31 + tag.charCodeAt(i)) >>> 0;
  }
  return `tag-c${hash % 8}`;
}

export function TagsInput({ value, onChange, disabled }: TagsInputProps) {
  const [input, setInput] = useState("");

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter") {
      e.preventDefault();
      const tag = input.trim();
      if (!tag || value.includes(tag)) return;
      onChange([...value, tag]);
      setInput("");
    }
    if (e.key === "Backspace" && input === "" && value.length > 0) {
      onChange(value.slice(0, -1));
    }
  }

  function removeTag(tag: string) {
    onChange(value.filter((t) => t !== tag));
  }

  return (
    <div className="tags-input">
      {value.map((tag) => (
        <span key={tag} className={`tags-input__chip ${tagColorClass(tag)}`}>
          <span className="tags-input__chip-text">{tag}</span>
          {!disabled && (
            <button
              type="button"
              className="tags-input__remove"
              aria-label={`Remove ${tag}`}
              onClick={() => removeTag(tag)}
            >
              <svg width="8" height="8" viewBox="0 0 8 8" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" aria-hidden="true">
                <line x1="1" y1="1" x2="7" y2="7" /><line x1="7" y1="1" x2="1" y2="7" />
              </svg>
            </button>
          )}
        </span>
      ))}
      {!disabled && (
        <input
          className="tags-input__input"
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={value.length === 0 ? "Add a tag…" : ""}
        />
      )}
    </div>
  );
}

export { tagColorClass };
