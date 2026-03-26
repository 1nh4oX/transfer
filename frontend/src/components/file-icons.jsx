const FileIcons = {
  PDF: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#DE7575]">
      <path d="M7 18H17V16H7V18ZM7 14H17V12H7V14ZM19 8L13 2H6C4.9 2 4.01 2.9 4.01 4L4 20C4 21.1 4.89 22 5.99 22H18C19.1 22 20 21.1 20 20V8ZM18 20H6V4H12V9H17V20Z" fill="currentColor" />
    </svg>
  ),
  Word: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#6C88C4]">
      <path d="M14 2H6C4.9 2 4 2.9 4 4V20C4 21.1 4.9 22 6 22H18C19.1 22 20 21.1 20 20V8L14 2ZM15.2 18H13.6L12 13.8L10.4 18H8.8L6.6 11H8.3L9.6 15.8L11.2 11H12.8L14.4 15.8L15.7 11H17.4L15.2 18ZM13 9V3.5L18.5 9H13Z" fill="currentColor" />
    </svg>
  ),
  Excel: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#6DB285]">
      <path d="M14 2H6C4.9 2 4.01 2.9 4.01 4L4 20C4 21.1 4.89 22 5.99 22H18C19.1 22 20 21.1 20 20V8L14 2ZM17.2 18H15.5L13.3 14.2L11.1 18H9.4L12.4 13L9.6 8H11.3L13.3 11.6L15.3 8H17L14.2 13L17.2 18ZM13 9V3.5L18.5 9H13Z" fill="currentColor" />
    </svg>
  ),
  PPT: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#D98E5F]">
      <path d="M14 2H6C4.9 2 4 2.9 4 4V20C4 21.1 4.9 22 6 22H18C19.1 22 20 21.1 20 20V8L14 2ZM11 18H9V11H11.5C13.43 11 15 12.57 15 14.5C15 16.43 13.43 18 11.5 18H11ZM13 9V3.5L18.5 9H13ZM11 13H12C12.55 13 13 13.45 13 14C13 14.55 12.55 15 12 15H11V13Z" fill="currentColor" />
    </svg>
  ),
  Image: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#7B9EBB]">
      <path d="M21 19V5C21 3.9 20.1 3 19 3H5C3.9 3 3 3.9 3 5V19C3 20.1 3.9 21 5 21H19C20.1 21 21 20.1 21 19ZM8.5 13.5L11 16.51L14.5 12L19 18H5L8.5 13.5Z" fill="currentColor" />
    </svg>
  ),
  Video: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#9B82C3]">
      <path d="M18 3H6C4.9 3 4 3.9 4 5V19C4 20.1 4.9 21 6 21H18C19.1 21 20 20.1 20 19V5C20 3.9 19.1 3 18 3ZM10 16V8L15 12L10 16Z" fill="currentColor" />
    </svg>
  ),
  Audio: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#5B9CA1]">
      <path d="M12 3V12.26C11.5 12.09 11.02 12 10.5 12C8.01 12 6 14.01 6 16.5C6 18.99 8.01 21 10.5 21C12.99 21 15 18.99 15 16.5V6H19V3H12Z" fill="currentColor" />
    </svg>
  ),
  Archive: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#A1887F]">
      <path d="M20 6H4V20H20V6ZM14 15H10V13H14V15ZM14 11H10V9H14V11ZM12 8H10V6H12V8ZM14 8H12V6H14V8ZM14 4H10V2H14V4Z" fill="currentColor" />
    </svg>
  ),
  Code: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#738290]">
      <path d="M9.4 16.6L4.8 12L9.4 7.4L8 6L2 12L8 18L9.4 16.6ZM14.6 16.6L19.2 12L14.6 7.4L16 6L22 12L16 18L14.6 16.6Z" fill="currentColor" />
    </svg>
  ),
  Text: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-[#8C9BA5]">
      <path d="M14 2H6C4.9 2 4 2.9 4 4V20C4 21.1 4.9 22 6 22H18C19.1 22 20 21.1 20 20V8L14 2ZM16 18H8V16H16V18ZM16 14H8V12H16V14ZM13 9V3.5L18.5 9H13Z" fill="currentColor" />
    </svg>
  ),
  Generic: () => (
    <svg viewBox="0 0 24 24" fill="none" className="w-full h-full text-slate-300">
      <path d="M14 2H6C4.89543 2 4 2.89543 4 4V20C4 21.1046 4.89543 22 6 22H18C19.1046 22 20 21.1046 20 20V8L14 2ZM13 9V3.5L18.5 9H13Z" fill="currentColor" />
      <circle cx="12" cy="16" r="1.5" fill="white" />
      <path d="M12 11V14" stroke="white" strokeWidth="2" strokeLinecap="round" />
    </svg>
  ),
}

export function getFileIcon(file) {
  const name = (file?.name || '').toLowerCase()
  const mime = (file?.mimeType || '').toLowerCase()

  if (name.endsWith('.pdf') || mime.includes('pdf')) return FileIcons.PDF
  if (/(\\.doc|\\.docx)$/.test(name)) return FileIcons.Word
  if (/(\\.xls|\\.xlsx)$/.test(name)) return FileIcons.Excel
  if (/(\\.ppt|\\.pptx)$/.test(name)) return FileIcons.PPT
  if (mime.startsWith('image/') || /\\.(png|jpe?g|gif|webp|bmp|svg)$/.test(name)) return FileIcons.Image
  if (mime.startsWith('video/') || /\\.(mp4|mov|avi|mkv|webm)$/.test(name)) return FileIcons.Video
  if (mime.startsWith('audio/') || /\\.(mp3|wav|m4a|flac|ogg)$/.test(name)) return FileIcons.Audio
  if (/\\.(zip|rar|7z|tar|gz)$/.test(name)) return FileIcons.Archive
  if (/\\.(js|ts|jsx|tsx|java|go|py|rs|c|cpp|h|hpp|json|yaml|yml|xml|html|css|sql|sh)$/.test(name)) return FileIcons.Code
  if (mime.startsWith('text/') || /\\.(txt|md|log)$/.test(name)) return FileIcons.Text
  return FileIcons.Generic
}
