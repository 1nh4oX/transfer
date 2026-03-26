import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import {
  AlertCircle,
  CheckCircle2,
  Clock,
  Copy,
  Download,
  FileBox,
  Folder,
  LogOut,
  Search,
  Send,
  Share2,
  Trash2,
  UploadCloud,
  X,
} from 'lucide-react'
import { getFileIcon } from './components/file-icons'
import { findNodeByFolderID, flattenTree, formatBytes, formatTime, inferKind } from './lib/helpers'

const GLASS_PANEL = 'bg-white/60 backdrop-blur-xl border border-white/60 shadow-[0_8px_30px_rgb(0,0,0,0.04)]'
const SPRING_TRANSITION = 'transition-all duration-300 ease-out'
const TOKEN_KEY = 'cloud_token'
const AUTH_EXPIRED_EVENT = 'auth-expired'
const API_BASE = import.meta.env.VITE_API_BASE || ''
const USER_QUOTA_BYTES = 2 * 1024 * 1024 * 1024

async function apiFetch(path, options = {}, token) {
  const headers = new Headers(options.headers || {})
  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers })
  const contentType = res.headers.get('content-type') || ''
  const body = contentType.includes('application/json') ? await res.json().catch(() => ({})) : null

  if (!res.ok) {
    if (res.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
      window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
    }
    const msg = body?.message || `Request failed: ${res.status}`
    throw new Error(msg)
  }
  return body
}

function Toast({ message, type }) {
  if (!message) return null
  return (
    <div
      className={`fixed top-6 left-1/2 -translate-x-1/2 z-50 px-6 py-3 rounded-full flex items-center gap-2 shadow-lg border ${
        type === 'error' ? 'bg-red-50 border-red-200 text-red-600' : 'bg-slate-800 text-white border-slate-700'
      }`}
    >
      {type === 'error' ? <AlertCircle size={16} /> : <CheckCircle2 size={16} />}
      <span className="text-sm font-medium">{message}</span>
    </div>
  )
}

function LoginPage({ onLogin, showToast }) {
  const [mode, setMode] = useState('login')
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [errorMessage, setErrorMessage] = useState('')

  const submit = async (e) => {
    e.preventDefault()
    setErrorMessage('')
    if (mode === 'register' && password !== confirmPassword) {
      setErrorMessage('两次密码不一致')
      showToast('两次密码不一致', 'error')
      return
    }
    setLoading(true)
    try {
      const path = mode === 'register' ? '/api/auth/register' : '/api/auth/login'
      const data = await apiFetch(path, {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      })
      onLogin(data.token)
      showToast(mode === 'register' ? '注册并登录成功' : '登录成功')
    } catch (err) {
      setErrorMessage(err.message || '登录/注册失败')
      showToast(err.message, 'error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-[#f4f5f7] flex items-center justify-center p-4">
      <div className={`relative w-full max-w-md ${GLASS_PANEL} p-10 rounded-[2rem]`}>
        <div className="mb-8 text-center">
          <div className="w-12 h-12 bg-slate-800 rounded-2xl flex items-center justify-center text-white mx-auto mb-4">
            <FileBox size={24} />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-800">登录 CloudSpace</h1>
          <p className="text-sm text-slate-500 mt-2">输入账号密码即可登录或注册</p>
        </div>

        <div className="mb-4 flex bg-slate-100 rounded-xl p-1">
          <button
            type="button"
            onClick={() => setMode('login')}
            className={`flex-1 py-2 text-sm rounded-lg ${mode === 'login' ? 'bg-white shadow-sm' : 'text-slate-500'}`}
          >
            登录
          </button>
          <button
            type="button"
            onClick={() => setMode('register')}
            className={`flex-1 py-2 text-sm rounded-lg ${mode === 'register' ? 'bg-white shadow-sm' : 'text-slate-500'}`}
          >
            注册
          </button>
        </div>

        <form onSubmit={submit} className="space-y-5">
          {errorMessage && (
            <div className="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">{errorMessage}</div>
          )}
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="w-full bg-white/70 border border-white/80 px-4 py-3 rounded-xl text-sm focus:outline-none"
            placeholder="用户名"
          />
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full bg-white/70 border border-white/80 px-4 py-3 rounded-xl text-sm focus:outline-none"
            placeholder="密码"
          />
          {mode === 'register' && (
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="w-full bg-white/70 border border-white/80 px-4 py-3 rounded-xl text-sm focus:outline-none"
              placeholder="确认密码"
            />
          )}
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-slate-800 text-white font-medium py-3 rounded-xl disabled:opacity-50"
          >
            {loading ? (mode === 'register' ? '注册中...' : '登录中...') : mode === 'register' ? '注册并进入' : '进入空间'}
          </button>
        </form>
      </div>
    </div>
  )
}

function Sidebar({ tree, currentFolderID, onSelectFolder }) {
  const renderNodes = (nodes, depth = 0) => {
    return (nodes || []).map((node) => {
      const active = currentFolderID === node.folder.id
      return (
        <div key={node.folder.id}>
          <button
            onClick={() => onSelectFolder(node.folder.id)}
            className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm text-left ${
              active ? 'bg-white shadow-sm text-slate-800 font-medium' : 'text-slate-600 hover:bg-white/70'
            }`}
            style={{ paddingLeft: `${12 + depth * 16}px` }}
          >
            <Folder size={16} className="text-slate-600 shrink-0" />
            <span className="truncate">{node.folder.name}</span>
          </button>
          {node.children?.length > 0 && renderNodes(node.children, depth + 1)}
        </div>
      )
    })
  }

  return (
    <aside className={`hidden lg:flex w-64 flex-col ${GLASS_PANEL} rounded-3xl p-5 shrink-0`}>
      <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-4 px-2">目录</h3>
      <button
        onClick={() => onSelectFolder(null)}
        className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm mb-2 ${
          currentFolderID == null ? 'font-medium bg-white shadow-sm text-slate-800' : 'text-slate-600 hover:bg-white/70'
        }`}
      >
        <Folder size={18} className="text-slate-700" />
        根目录
      </button>
      <div className="flex-1 overflow-auto space-y-1 pr-1">{renderNodes(tree)}</div>
    </aside>
  )
}

function PreviewModal({ file, onClose, token }) {
  const kind = inferKind(file.name, file.mimeType || '')
  const [textContent, setTextContent] = useState('')
  const [blobUrl, setBlobUrl] = useState('')
  const [officeSrc, setOfficeSrc] = useState('')
  const [officeDownloadUrl, setOfficeDownloadUrl] = useState('')
  const [officePreviewError, setOfficePreviewError] = useState(false)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    let revoked = ''
    const run = async () => {
      if (!file?.id) return
      if (kind !== 'image' && kind !== 'pdf' && kind !== 'text' && kind !== 'office') return
      setOfficePreviewError(false)
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
          const officeUrl = `https://view.officeapps.live.com/op/embed.aspx?src=${encodeURIComponent(publicDownloadUrl)}`
          setOfficeDownloadUrl(publicDownloadUrl)
          setOfficeSrc(officeUrl)
        } else {
          const res = await fetch(`${API_BASE}/api/files/${file.id}/download`, {
            headers: { Authorization: `Bearer ${token}` },
          })
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
        }
        setTextContent('预览暂不可用，请直接下载。')
      } finally {
        setLoading(false)
      }
    }
    run()
    return () => {
      if (revoked) URL.revokeObjectURL(revoked)
    }
  }, [file, kind, token])

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
          {!loading && kind === 'pdf' && blobUrl && (
            <iframe title="pdf" src={blobUrl} className="w-full h-[70vh] border rounded-xl bg-white" />
          )}
          {!loading && kind === 'text' && (
            <pre className="w-full bg-white p-4 rounded-xl text-sm text-slate-700 overflow-auto border">{textContent}</pre>
          )}
          {!loading && kind === 'office' && officeSrc && !officePreviewError && (
            <iframe
              title="office"
              src={officeSrc}
              className="w-full h-[70vh] border rounded-xl bg-white"
              onError={() => setOfficePreviewError(true)}
            />
          )}
          {!loading && kind === 'office' && officePreviewError && (
            <div className="space-y-3">
              <p className="text-slate-500">Word/Excel/PPT 在线预览失败，请直接下载查看。</p>
              <a
                href={officeDownloadUrl || '#'}
                className="inline-flex items-center gap-2 bg-slate-800 text-white px-4 py-2 rounded-xl"
              >
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

function ShareModal({ file, onClose, token, showToast }) {
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
  }, [file.id, showToast, token])

  const copyUrl = async () => {
    if (!url) return
    try {
      await navigator.clipboard.writeText(url)
      showToast('链接已复制')
      onClose()
    } catch {
      showToast('复制失败，请手动复制', 'error')
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
            <button
              onClick={copyUrl}
              className="mt-4 w-full flex items-center justify-center gap-2 bg-slate-800 text-white py-3 rounded-xl font-medium"
            >
              <Copy size={16} /> 复制链接
            </button>
          </>
        )}
      </div>
    </div>,
    document.body,
  )
}

function FileArea({ token, showToast, tree, currentFolderID, onSelectFolder, refreshTree, totalUsedBytes = 0 }) {
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
        if (currentFolderID) {
          formData.append('folderId', currentFolderID)
        }
        try {
          const item = await apiFetch('/api/files/upload', { method: 'POST', body: formData }, token)
          uploaded.push(item)
        } catch {
          failed += 1
        }
      }

      if (uploaded.length > 0) {
        setFiles((prev) => [...uploaded, ...prev])
      }
      if (uploaded.length > 0) {
        refreshTree()
      }
      if (failed === 0) {
        showToast(`已上传 ${uploaded.length} 个文件`)
      } else {
        showToast(`上传完成：成功 ${uploaded.length}，失败 ${failed}`, failed > 0 ? 'error' : 'success')
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
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  return (
    <section className={`flex-1 flex flex-col relative ${GLASS_PANEL} rounded-3xl overflow-hidden`}>
      <div className="flex items-center justify-between p-6 border-b border-white/50">
        <div>
          <h2 className="text-xl font-bold tracking-tight">我的文件</h2>
          <p className="text-xs text-slate-500 mt-1">{currentFolderName}</p>
          <div className="mt-3 w-56">
            <div className="flex items-center justify-between text-[11px] text-slate-500 mb-1">
              <span>空间使用</span>
              <span>{formatBytes(totalUsedBytes)} / 2.00 GB</span>
            </div>
            <div className="h-2 rounded-full bg-slate-200/70 overflow-hidden">
              <div
                className={`h-full ${usagePercent > 90 ? 'bg-red-500' : 'bg-slate-700'}`}
                style={{ width: `${usagePercent}%` }}
              ></div>
            </div>
          </div>
        </div>
        <div className="relative flex items-center gap-2">
          <button
            onClick={createFolder}
            className="flex items-center gap-2 bg-white text-slate-700 px-4 py-2.5 rounded-full text-sm font-medium border border-slate-200"
          >
            <Folder size={16} />
            <span>新建文件夹</span>
          </button>
          <input type="file" id="file-upload" className="hidden" multiple onChange={upload} />
          <label htmlFor="file-upload" className="flex items-center gap-2 bg-slate-800 text-white px-5 py-2.5 rounded-full text-sm font-medium cursor-pointer">
            <UploadCloud size={16} />
            <span>{uploading ? '上传中...' : '批量上传'}</span>
          </label>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6">
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
            {files.map((file) => (
              (() => {
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
              })()
            ))}
          </div>
        )}
      </div>

      {previewFile && <PreviewModal file={previewFile} token={token} onClose={() => setPreviewFile(null)} />}
      {shareFile && <ShareModal file={shareFile} token={token} onClose={() => setShareFile(null)} showToast={showToast} />}
    </section>
  )
}

function NotesArea({ token, showToast }) {
  const [notes, setNotes] = useState([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const scrollRef = useRef(null)

  const loadNotes = async () => {
    try {
      const data = await apiFetch('/api/notes?page=1&pageSize=100', {}, token)
      const items = data.items || []
      setNotes(items.slice().reverse())
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  useEffect(() => {
    loadNotes()
  }, [])

  useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight
  }, [notes])

  const send = async () => {
    const content = input.trim()
    if (!content) return
    setLoading(true)
    setInput('')
    try {
      const item = await apiFetch('/api/notes', { method: 'POST', body: JSON.stringify({ content }) }, token)
      setNotes((prev) => [...prev, item])
    } catch (err) {
      showToast(err.message, 'error')
      setInput(content)
    } finally {
      setLoading(false)
    }
  }

  return (
    <aside className={`w-[360px] h-full min-h-0 flex flex-col ${GLASS_PANEL} rounded-3xl overflow-hidden shrink-0`}>
      <div className="p-4 border-b border-white/50 bg-white/30 flex items-center justify-between">
        <div>
          <h3 className="font-semibold text-sm text-slate-800 tracking-wide">速记对话框</h3>
          <p className="text-[10px] text-slate-500">仅自己可见</p>
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto overscroll-contain p-4 space-y-5 flex flex-col bg-slate-50/30" ref={scrollRef}>
        {notes.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center text-slate-400 mt-10 opacity-60">
            <Clock size={32} className="mb-2 text-slate-300" />
            <span className="text-xs">在这里随手记录想法或待办...</span>
          </div>
        ) : (
          notes.map((note) => (
            <div key={note.id} className="flex flex-col items-end">
              <div className="bg-slate-800 text-white p-3.5 rounded-2xl rounded-br-sm shadow-md max-w-[85%]">
                <p className="text-sm leading-relaxed whitespace-pre-wrap break-words">{note.content}</p>
              </div>
              <span className="text-[10px] text-slate-400 mt-1.5 mr-1 font-medium">{formatTime(note.createdAt)}</span>
            </div>
          ))
        )}
        {loading && <div className="text-xs text-slate-400 text-right">发送中...</div>}
      </div>

      <div className="p-4 bg-white/40 border-t border-white/50">
        <div className="relative flex items-end bg-white/80 border border-slate-200/60 rounded-[20px] shadow-sm">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            placeholder="随时记录..."
            className="w-full bg-transparent rounded-[20px] pl-4 pr-12 py-3 text-sm focus:outline-none resize-none h-[44px]"
            rows="1"
          />
          <button
            onClick={send}
            disabled={!input.trim()}
            className={`absolute right-2 bottom-2 p-2 rounded-xl flex items-center justify-center ${
              input.trim() ? 'bg-slate-800 text-white' : 'bg-slate-100 text-slate-400 cursor-not-allowed'
            }`}
          >
            <Send size={16} />
          </button>
        </div>
      </div>
    </aside>
  )
}

export default function App() {
  const [token, setToken] = useState(localStorage.getItem(TOKEN_KEY) || '')
  const [toast, setToast] = useState({ message: '', type: '' })
  const [tree, setTree] = useState([])
  const [currentFolderID, setCurrentFolderID] = useState(null)
  const [totalUsedBytes, setTotalUsedBytes] = useState(0)

  const showToast = (message, type = 'success') => {
    setToast({ message, type })
    window.clearTimeout(window.__toastTimer)
    window.__toastTimer = window.setTimeout(() => setToast({ message: '', type: '' }), 3000)
  }

  const logout = () => {
    localStorage.removeItem(TOKEN_KEY)
    setToken('')
  }

  const setLoggedIn = (nextToken) => {
    localStorage.setItem(TOKEN_KEY, nextToken)
    setToken(nextToken)
    setCurrentFolderID(null)
  }

  const loadTree = async () => {
    if (!token) return
    try {
      const data = await apiFetch('/api/tree', {}, token)
      setTree(data.rootFolders || [])
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  useEffect(() => {
    loadTree()
  }, [token])

  useEffect(() => {
    const onAuthExpired = () => {
      if (token) {
        setToken('')
        showToast('登录已失效，请重新登录', 'error')
      }
    }
    window.addEventListener(AUTH_EXPIRED_EVENT, onAuthExpired)
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, onAuthExpired)
  }, [token])

  const loadTotalUsage = async () => {
    if (!token) return
    try {
      const folderIDs = [null, ...flattenTree(tree).map((f) => f.id)]
      let sum = 0
      for (const fid of folderIDs) {
        let page = 1
        while (true) {
          const query = fid
            ? `/api/files?page=${page}&pageSize=100&folderId=${encodeURIComponent(fid)}`
            : `/api/files?page=${page}&pageSize=100`
          const data = await apiFetch(query, {}, token)
          const items = data.items || []
          sum += items.reduce((acc, item) => acc + (item.size || 0), 0)
          const total = data.total || 0
          if (page * 100 >= total || items.length === 0) break
          page += 1
        }
      }
      setTotalUsedBytes(sum)
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  useEffect(() => {
    loadTotalUsage()
  }, [token, tree])

  if (!token) {
    return (
      <div className="min-h-screen bg-[#f4f5f7]">
        <Toast message={toast.message} type={toast.type} />
        <LoginPage onLogin={setLoggedIn} showToast={showToast} />
      </div>
    )
  }

  return (
    <div className="h-screen bg-[#f4f5f7] text-slate-800 font-sans overflow-x-auto overflow-y-hidden flex flex-col">
      <Toast message={toast.message} type={toast.type} />
      <header className={`h-16 ${GLASS_PANEL} rounded-b-2xl mx-4 mt-2 flex items-center justify-between px-6 z-10 shrink-0`}>
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-slate-800 flex items-center justify-center text-white shadow-sm">
            <FileBox size={18} />
          </div>
          <span className="font-semibold text-lg tracking-tight">CloudSpace</span>
        </div>
        <button onClick={logout} className={`p-2 rounded-full hover:bg-slate-200/50 text-slate-500 hover:text-slate-800 ${SPRING_TRANSITION}`}>
          <LogOut size={18} />
        </button>
      </header>

      <main className="flex-1 min-h-0 flex overflow-hidden p-4 gap-4 max-w-[1600px] mx-auto w-full min-w-[1320px]">
        <Sidebar tree={tree} currentFolderID={currentFolderID} onSelectFolder={setCurrentFolderID} />
        <FileArea
          token={token}
          showToast={showToast}
          tree={tree}
          currentFolderID={currentFolderID}
          onSelectFolder={setCurrentFolderID}
          refreshTree={loadTree}
          totalUsedBytes={totalUsedBytes}
        />
        <NotesArea token={token} showToast={showToast} />
      </main>
    </div>
  )
}
