import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(duration)
dayjs.extend(relativeTime)

export function formatTimeDifference(pushTime: Date, deadlineTime: Date): string {
  const pushDayjs = dayjs(pushTime)
  const deadlineDayjs = dayjs(deadlineTime)

  const diff = deadlineDayjs.diff(pushDayjs)
  const duration = dayjs.duration(diff)

  if (diff > 0) {
    // Push was before deadline
    if (duration.asDays() >= 1) {
      return `${Math.floor(duration.asDays())} days before deadline`
    } else if (duration.asHours() >= 1) {
      return `${Math.floor(duration.asHours())} hours before deadline`
    } else {
      return `${Math.floor(duration.asMinutes())} minutes before deadline`
    }
  } else {
    // Push was after deadline
    const absDuration = dayjs.duration(Math.abs(diff))
    if (absDuration.asDays() >= 1) {
      return `${Math.floor(absDuration.asDays())} days after deadline`
    } else if (absDuration.asHours() >= 1) {
      return `${Math.floor(absDuration.asHours())} hours after deadline`
    } else {
      return `${Math.floor(absDuration.asMinutes())} minutes after deadline`
    }
  }
}

export function formatDateTime(dateString: string): string {
  return dayjs(dateString).format('YYYY-MM-DD HH:mm:ss UTC')
}

export function isValidDateTime(dateString: string): boolean {
  return dayjs(dateString).isValid()
}