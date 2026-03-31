import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { Copy, Download, Folder, Search, Share2, Trash2, UploadCloud, X } from 'lucide-react'
import { getFileIcon } from './file-icons'
import { findNodeByFolderID, flattenTree, formatBytes, formatTime, inferKind } from '../lib/helpers'

const GLASS_PANEL = 'bg-white/60 backdrop-blur-xl border border-white/60 shadow-[0_8px_30px_rgb(0,0,0,0.04)]'
const SPRING_TRANSITION = 'transition-all duration-300 ease-out'
const API_BASE = import.meta.env.VITE_API_BASE || ''
const USER_QUOTA_BYTES = 2 * 1024 * 1024 * 1024

function PreviewModal({ file, onClose, token, apiFetch }) {
  const kind = inferKind(file.name, file.mimeType || '')
  const [textContent, setTextContent] = useState('')
  const [blobUrl, setBlobUrl] = useState('')
  const [officeSrc, setOfficeSrc] = useState('')
  const [officeDownloadUrl, setOfficeDownloadUrl] = useState('')
  const [officePreviewError, setOfficePreviewError] = useState(false)
  const [officePreviewMessage, setOfficePreviewMessage] = useState('')
  const [officeLoaded, setOfficeLoaded] = useState(false)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    let revoked = ''
    let officeTimeout = 0
    const run = async () => {
      if (!file?.id) return
      if (kind !== 'image' && kind !== 'pdf' && kind !== 'text' && kind !== 'office') return
      setOfficePreviewError(false)
      setOfficePreviewMessage('')
      setOfficeLoaded(false)
      setOfficeSrc('')
      setOfficeDownloadUrl('')
      setLoading(true)
      try {
        if (kind === 'office') {
          const share = await apiFetch(
            '/api/shares',
            {
              method: 'POST',
              body: JSON.stringify({ itemType: 'file', itemId: file.id }),
            },
            token,
          )
          const publicDownloadUrl = `${share.shareUrl}/download`
          if (publicDownloadUrl.startsWith('http://')) {
            setOfficePreviewError(true)
            setOfficePreviewMessage('Office 在线预览通常需要 HTTPS 公网链接，当前为 HTTP，请直接下载查看。')
            setOfficeDownloadUrl(publicDownloadUrl)
            return
          }

          const officeUrl = `https://view.officeapps.live.com/op/embed.aspx?src=${encodeURIComponent(publicDownloadUrl)}`
          setOfficeDownloadUrl(publicDownloadUrl)
          setOfficeSrc(officeUrl)
          officeTimeout = window.setTimeout(() => {
            setOfficePreviewError(true)
            setOfficePreviewMessage('预览服务响应超时，请直接下载查看。')
          }, 12000)
        } else {
          const ctrl = new AbortController()
          const timeout = window.setTimeout(() => ctrl.abort(), 12000)
          const res = await fetch(`${API_BASE}/api/files/${file.id}/download`, {
            headers: { Authorization: `Bearer ${token}` },
            signal: ctrl.signal,
          })
          window.clearTimeout(timeout)
          if (!res.ok) throw new Error('加载预览失败')
          if (kind === 'text') {
            setTextContent(await res.text())
          } else {
            const blob = await res.blob()
            const u = URL.createObjectURL(blob)
            revoked = u
            setBlobUrl(u)
          }
        }
      } catch {
        if (kind === 'office') {
          setOfficePreviewError(true)
          setOfficePreviewMessage('Office 在线预览失败，请直接下载查看。')
        } else {
          setTextContent('预览暂不可用，请直接下载。')
        }
      } finally {
        setLoading(false)
      }
    }

    run()
    return () => {
      if (officeTimeout) window.clearTimeout(officeTimeout)
      if (revoked) URL.revokeObjectURL(revoked)
    }
  }, [file, kind, token, apiFetch])

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
      <div className="absolute inset-0 bg-slate-900/25 backdrop-blur-sm" onClick={onClose}></div>
      <div className="relative w-full max-w-4xl max-h-full bg-white rounded-3xl shadow-2xl overflow-hidden flex flex-col">
        <div className="flex items-center justify-between p-4 border-b border-slate-100">
          <span className="font-medium truncate">{file.name}</span>
          <button onClick={onClose} className="p-2 rounded-full hover:bg-slate-100 text-slate-500">
            <X size={18} />
          </button>
        </div>

        <div className="flex-1 overflow-auto p-6 min-h-[400px] bg-slate-50/50">
          {loading && <p className="text-slate-500">加载中...</p>}
          {!loading && kind === 'image' && blobUrl && <img src={blobUrl} alt={file.name} className="max-h-[70vh] mx-auto" />}
          {!loading && kind === 'pdf' && blobUrl && <iframe title="pdf" src={blobUrl} className="w-full h-[70vh] border rounded-xl bg-white" />}
          {!loading && kind === 'text' && (
            <pre className="w-full bg-white p-4 rounded-xl text-sm text-slate-700 overflow-auto border">{textContent}</pre>
          )}
          {!loading && kind === 'office' && officeSrc && !officePreviewError && (
            <iframe
              title="office"
              src={officeSrc}
              className="w-full h-[70vh] border rounded-xl bg-white"
              onLoad={() => setOfficeLoaded(true)}
              onError={() => {
                setOfficePreviewError(true)
                setOfficePreviewMessage('Office 在线预览失败，请直接下载查看。')
              }}
            />
          )}
          {!loading && kind === 'office' && officeSrc && !officePreviewError && !officeLoaded && (
            <p className="text-slate-500 mt-2">正在加载 Office 预览...</p>
          )}
          {!loading && kind === 'office' && officePreviewError && (
            <div className="space-y-3">
              <p className="text-slate-500">{officePreviewMessage || 'Word/Excel/PPT 在线预览失败，请直接下载查看。'}</p>
              <a href={officeDownloadUrl || '#'} className="inline-flex items-center gap-2 bg-slate-800 text-white px-4 py-2 rounded-xl">
                <Download size={16} /> 下载文件
              </a>
            </div>
          )}
          {!loading && kind === 'file' && <p className="text-slate-500">该文件类型暂不支持预览。</p>}
        </div>
      </div>
    </div>,
    document.body,
  )
}

function ShareModal({ file, onClose, token, showToast, apiFetch }) {
  const [url, setUrl] = useState('')
  const [metaUrl, setMetaUrl] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const createShare = async () => {
      setLoading(true)
      try {
        const data = await apiFetch(
          '/api/shares',
          {
            method: 'POST',
            body: JSON.stringify({ itemType: 'file', itemId: file.id }),
          },
          token,
        )
        setMetaUrl(data.shareUrl)
        setUrl(`${data.shareUrl}/download`)
      } catch (err) {
        showToast(err.message, 'error')
      } finally {
        setLoading(false)
      }
    }
    createShare()
  }, [file.id, showToast, token, apiFetch])

  const copyUrl = async () => {
    if (!url) return
    const fallbackCopy = () => {
      const textarea = document.createElement('textarea')
      textarea.value = url
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.focus()
      textarea.select()
      const ok = document.execCommand('copy')
      document.body.removeChild(textarea)
      return ok
    }

    try {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(url)
      } else if (!fallbackCopy()) {
        throw new Error('copy-failed')
      }
      showToast('链接已复制')
      onClose()
    } catch {
      if (fallbackCopy()) {
        showToast('链接已复制')
        onClose()
      } else {
        showToast('复制失败，请手动复制', 'error')
      }
    }
  }

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-slate-900/20 backdrop-blur-sm" onClick={onClose}></div>
      <div className="relative w-full max-w-sm bg-white rounded-3xl shadow-2xl border border-white p-6">
        <div className="flex justify-between items-start mb-6">
          <div className="w-10 h-10 bg-slate-100 rounded-full flex items-center justify-center text-slate-600">
            <Share2 size={20} />
          </div>
          <button onClick={onClose} className="p-1.5 rounded-full hover:bg-slate-100 text-slate-400">
            <X size={16} />
          </button>
        </div>
        <h3 className="text-lg font-bold mb-1">分享文件</h3>
        <p className="text-sm text-slate-500 mb-4 truncate">{file.name}</p>
        {loading ? (
          <div className="h-10 bg-slate-100 rounded-xl animate-pulse w-full"></div>
        ) : (
          <>
            <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl text-sm text-slate-600 font-mono break-all">{url}</div>
            <p className="mt-2 text-xs text-slate-500 break-all">信息地址：{metaUrl}</p>
            <button onClick={copyUrl} className="mt-4 w-full flex items-center justify-center gap-2 bg-slate-800 text-white py-3 rounded-xl font-medium">
              <Copy size={16} /> 复制链接
            </button>
          </>
        )}
      </div>
    </div>,
    document.body,
  )
}

export default function FileArea({ token, showToast, tree, currentFolderID, onSelectFolder, refreshTree, totalUsedBytes = 0, mobile = false, apiFetch }) {
  const [files, setFiles] = useState([])
  const [loading, setLoading] = useState(true)
  const [uploading, setUploading] = useState(false)
  const [previewFile, setPreviewFile] = useState(null)
  const [shareFile, setShareFile] = useState(null)

  const allFolders = flattenTree(tree)
  const currentNode = currentFolderID ? findNodeByFolderID(tree, currentFolderID) : null
  const currentFolderName = currentNode?.folder?.name || 'Root'
  const childFolders = currentFolderID == null ? tree.map((n) => n.folder) : (currentNode?.children || []).map((n) => n.folder)
  const usagePercent = Math.min(100, Math.round((totalUsedBytes / USER_QUOTA_BYTES) * 100))

  const loadFiles = async () => {
    setLoading(true)
    try {
      const query = currentFolderID
        ? `/api/files?page=1&pageSize=100&folderId=${encodeURIComponent(currentFolderID)}`
        : '/api/files?page=1&pageSize=100'
      const data = await apiFetch(query, {}, token)
      setFiles(data.items || [])
    } catch (err) {
      showToast(err.message, 'error')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadFiles()
  }, [currentFolderID])

  const upload = async (event) => {
    const selected = Array.from(event.target.files || [])
    if (selected.length === 0) return

    const batchBytes = selected.reduce((sum, file) => sum + file.size, 0)
    if (totalUsedBytes + batchBytes > USER_QUOTA_BYTES) {
      showToast('超出 2GB 配额，请先删除部分文件再上传', 'error')
      event.target.value = ''
      return
    }

    setUploading(true)
    const uploaded = []
    let failed = 0

    try {
      for (const file of selected) {
        const formData = new FormData()
        formData.append('file', file)
        if (currentFolderID) formData.append('folderId', currentFolderID)

        try {
          const item = await apiFetch('/api/files/upload', { method: 'POST', body: formData }, token)
          uploaded.push(item)
        } catch (err) {
          if (String(err?.message || '').includes('413')) {
            showToast('上传失败：文件过大或网关限制（请检查 Nginx client_max_body_size）', 'error')
            break
          }
          failed += 1
        }
      }

      if (uploaded.length > 0) {
        setFiles((prev) => [...uploaded, ...prev])
        refreshTree()
      }
      if (failed === 0) {
        showToast(`已上传 ${uploaded.length} 个文件`)
      } else {
        showToast(`上传完成：成功 ${uploaded.length}，失败 ${failed}`, 'error')
      }
    } finally {
      setUploading(false)
      event.target.value = ''
    }
  }

  const createFolder = async () => {
    const name = window.prompt('新建文件夹名称：')
    if (!name || !name.trim()) return

    try {
      await apiFetch(
        '/api/folders',
        {
          method: 'POST',
          body: JSON.stringify({ name: name.trim(), parentId: currentFolderID ?? null }),
        },
        token,
      )
      showToast('文件夹创建成功')
      refreshTree()
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  const renameCurrentFolder = async () => {
    if (!currentFolderID) {
      showToast('根目录不可重命名', 'error')
      return
    }
    const oldName = currentNode?.folder?.name || ''
    const nextName = window.prompt('重命名文件夹：', oldName)
    if (!nextName || !nextName.trim() || nextName.trim() === oldName) return

    try {
      await apiFetch(
        `/api/folders/${currentFolderID}/rename`,
        {
          method: 'PATCH',
          body: JSON.stringify({ name: nextName.trim() }),
        },
        token,
      )
      showToast('文件夹已重命名')
      refreshTree()
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  const moveFile = async (fileID, targetFolderID) => {
    try {
      await apiFetch(
        `/api/files/${fileID}/move`,
        {
          method: 'PATCH',
          body: JSON.stringify({ targetFolderId: targetFolderID || null }),
        },
        token,
      )
      showToast('文件已移动')
      loadFiles()
      refreshTree()
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  const removeFile = async (id) => {
    try {
      await apiFetch(`/api/files/${id}`, { method: 'DELETE' }, token)
      setFiles((prev) => prev.filter((f) => f.id !== id))
      showToast('文件已删除')
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  const downloadFile = async (file) => {
    try {
      const res = await fetch(`${API_BASE}/api/files/${file.id}/download`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) throw new Error('下载失败')
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = file.name
      a.style.display = 'none'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (err) {
      try {
        const share = await apiFetch(
          '/api/shares',
          {
            method: 'POST',
            body: JSON.stringify({ itemType: 'file', itemId: file.id }),
          },
          token,
        )
        window.open(`${share.shareUrl}/download`, '_blank', 'noopener,noreferrer')
        showToast('已使用分享链接下载')
      } catch {
        showToast(err.message, 'error')
      }
    }
  }

  return (
    <section className={`${mobile ? 'h-full' : 'flex-1'} min-h-0 flex flex-col relative ${GLASS_PANEL} rounded-3xl overflow-hidden`}>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-4 sm:p-6 border-b border-white/50">
        <div>
          <h2 className="text-xl font-bold tracking-tight">我的文件</h2>
          <p className="text-xs text-slate-500 mt-1">{currentFolderName}</p>
          <div className="mt-3 w-56">
            <div className="flex items-center justify-between text-[11px] text-slate-500 mb-1">
              <span>空间使用</span>
              <span>{formatBytes(totalUsedBytes)} / 2.00 GB</span>
            </div>
            <div className="h-2 rounded-full bg-slate-200/70 overflow-hidden">
              <div className={`h-full ${usagePercent > 90 ? 'bg-red-500' : 'bg-slate-700'}`} style={{ width: `${usagePercent}%` }}></div>
            </div>
          </div>
        </div>

        <div className="relative flex items-center gap-2 flex-wrap">
          <button
            onClick={createFolder}
            className="flex items-center gap-2 bg-white text-slate-700 px-4 py-2.5 rounded-full text-sm font-medium border border-slate-200"
          >
            <Folder size={16} />
            <span>新建文件夹</span>
          </button>
          {currentFolderID && (
            <button
              onClick={renameCurrentFolder}
              className="flex items-center gap-2 bg-white text-slate-700 px-4 py-2.5 rounded-full text-sm font-medium border border-slate-200"
            >
              <span>重命名当前文件夹</span>
            </button>
          )}
          <input type="file" id={mobile ? 'file-upload-mobile' : 'file-upload'} className="hidden" multiple onChange={upload} />
          <label
            htmlFor={mobile ? 'file-upload-mobile' : 'file-upload'}
            className="flex items-center gap-2 bg-slate-800 text-white px-5 py-2.5 rounded-full text-sm font-medium cursor-pointer"
          >
            <UploadCloud size={16} />
            <span>{uploading ? '上传中...' : '批量上传'}</span>
          </label>
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto overscroll-y-contain [webkit-overflow-scrolling:touch] p-4 sm:p-6 [padding-bottom:calc(env(safe-area-inset-bottom)+1rem)]">
        {childFolders.length > 0 && (
          <div className="mb-4 flex flex-wrap gap-2">
            {childFolders.map((folder) => (
              <button
                key={folder.id}
                onClick={() => onSelectFolder(folder.id)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white border border-slate-200 text-slate-700 text-xs"
              >
                <Folder size={14} /> {folder.name}
              </button>
            ))}
          </div>
        )}

        {loading ? (
          <div className="text-slate-400">加载文件中...</div>
        ) : files.length === 0 ? (
          <div className="text-slate-400">暂无文件</div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {files.map((file) => {
              const Icon = getFileIcon(file)
              return (
                <div key={file.id} className={`group bg-white/40 border border-white/60 p-4 rounded-2xl ${SPRING_TRANSITION} hover:-translate-y-1`}>
                  <div className="flex items-start gap-4">
                    <div className="w-12 h-12 rounded-xl flex items-center justify-center shrink-0 border shadow-sm p-2.5 bg-slate-50 border-slate-200/60">
                      <Icon />
                    </div>
                    <div className="flex-1 min-w-0">
                      <h4 className="text-sm font-medium text-slate-800 break-all line-clamp-2">{file.name}</h4>
                      <div className="flex flex-wrap items-center gap-2 mt-1.5 text-xs text-slate-400">
                        <span>{formatBytes(file.size)}</span>
                        <span>{formatTime(file.createdAt)}</span>
                      </div>
                    </div>
                  </div>

                  <div className="mt-4 pt-3 border-t border-slate-200/60 flex items-center justify-end gap-2" onClick={(e) => e.stopPropagation()}>
                    <select
                      defaultValue={file.folderId || '__ROOT__'}
                      onChange={(e) => {
                        const v = e.target.value
                        moveFile(file.id, v === '__ROOT__' ? null : v)
                      }}
                      className="text-xs border border-slate-200 rounded-lg px-2 py-1 bg-white text-slate-600"
                      title="移动到目录"
                    >
                      <option value="__ROOT__">移动到根目录</option>
                      {allFolders.map((f) => (
                        <option key={f.id} value={f.id}>{`${' '.repeat(f.depth * 2)}${f.name}`}</option>
                      ))}
                    </select>
                    <button onClick={() => setPreviewFile(file)} className="p-2 bg-white rounded-full shadow-sm text-slate-500 hover:text-slate-800" title="预览">
                      <Search size={14} />
                    </button>
                    <button onClick={() => setShareFile(file)} className="p-2 bg-white rounded-full shadow-sm text-slate-500 hover:text-slate-800" title="分享">
                      <Share2 size={14} />
                    </button>
                    <button onClick={() => downloadFile(file)} className="p-2 bg-white rounded-full shadow-sm text-slate-500 hover:text-slate-800" title="下载">
                      <Download size={14} />
                    </button>
                    <button onClick={() => removeFile(file.id)} className="p-2 bg-white rounded-full shadow-sm text-red-400 hover:text-red-600" title="删除">
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>

      {previewFile && <PreviewModal file={previewFile} token={token} onClose={() => setPreviewFile(null)} apiFetch={apiFetch} />}
      {shareFile && <ShareModal file={shareFile} token={token} onClose={() => setShareFile(null)} showToast={showToast} apiFetch={apiFetch} />}
    </section>
  )
}
