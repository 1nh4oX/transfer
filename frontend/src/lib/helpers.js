export function formatBytes(bytes = 0) {
  if (bytes < 1024) return `${bytes} B`
  const units = ['KB', 'MB', 'GB', 'TB']
  let i = -1
  let val = bytes
  do {
    val /= 1024
    i += 1
  } while (val >= 1024 && i < units.length - 1)
  return `${val.toFixed(val >= 10 ? 1 : 2)} ${units[i]}`
}

export function formatTime(iso) {
  if (!iso) return '-'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function inferKind(name = '', mime = '') {
  const lower = name.toLowerCase()
  if (mime.startsWith('image/') || /\.(png|jpe?g|gif|webp|bmp)$/i.test(lower)) return 'image'
  if (mime.includes('pdf') || lower.endsWith('.pdf')) return 'pdf'
  if (mime.startsWith('text/') || /\.(txt|md|json|log)$/i.test(lower)) return 'text'
  if (/\.(doc|docx|xls|xlsx|ppt|pptx)$/i.test(lower)) return 'office'
  return 'file'
}

export function flattenTree(nodes, depth = 0, out = []) {
  for (const node of nodes || []) {
    out.push({ ...node.folder, depth })
    flattenTree(node.children, depth + 1, out)
  }
  return out
}

export function findNodeByFolderID(nodes, folderID) {
  for (const node of nodes || []) {
    if (node.folder.id === folderID) return node
    const inChild = findNodeByFolderID(node.children, folderID)
    if (inChild) return inChild
  }
  return null
}
