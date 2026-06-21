import MissingReport from '../components/MissingReport.jsx'
import ResultEditor from '../components/ResultEditor.jsx'
import LockSetting from '../components/LockSetting.jsx'
import PlayersCards from '../components/PlayersCards.jsx'

export default function Admin() {
  return (
    <div className="admin">
      <MissingReport />
      <ResultEditor />
      <LockSetting />
      <PlayersCards />
    </div>
  )
}
