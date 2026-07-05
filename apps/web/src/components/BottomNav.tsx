import { NavLink } from 'react-router-dom'
import { ClipboardList, Search, Package, Settings as SettingsIcon } from 'lucide-react'

const tabs = [
  { to: '/', label: 'Lists', icon: ClipboardList, end: true },
  { to: '/lookup', label: 'Lookup', icon: Search },
  { to: '/items', label: 'Items', icon: Package },
  { to: '/settings', label: 'Settings', icon: SettingsIcon }
]

export default function BottomNav() {
  return (
    <nav className="fixed inset-x-0 bottom-0 flex border-t border-gray-200 bg-white pb-[env(safe-area-inset-bottom)]">
      {tabs.map(({ to, label, icon: Icon, end }) => (
        <NavLink
          key={to}
          to={to}
          end={end}
          className={({ isActive }) =>
            `flex flex-1 flex-col items-center gap-0.5 py-2 text-xs ${
              isActive ? 'text-green-600' : 'text-gray-500'
            }`
          }
        >
          <Icon size={22} />
          {label}
        </NavLink>
      ))}
    </nav>
  )
}
