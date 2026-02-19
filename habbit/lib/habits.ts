import { Habit, HabitLog } from '@prisma/client';

export function isHabitDueToday(habit: Habit, logs: HabitLog[]) {
  const today = new Date();
  const day = today.getDay();
  const dateKey = today.toDateString();
  const already = logs.some(l => new Date(l.date).toDateString() === dateKey);

  if (habit.frequency === 'DAILY') {
    if (habit.daysOfWeek && habit.daysOfWeek.length > 0) {
      return habit.daysOfWeek.includes(day) && !already;
    }
    return !already;
  }

  if (habit.frequency === 'WEEKLY') {
    return !already; // weekly habits can be done any day until count reached (simplified M1)
  }

  if (habit.frequency === 'CUSTOM') {
    const last = logs.sort((a,b)=>+new Date(b.date)-+new Date(a.date))[0];
    if (!last) return true;
    const diffDays = Math.floor((Date.now() - new Date(last.date).getTime()) / (1000*60*60*24));
    return diffDays >= (habit.everyXDays || 1);
  }

  return true;
}

export function calcStreak(logs: HabitLog[]) {
  const dates = new Set(logs.map(l => new Date(l.date).toDateString()));
  let streak = 0;
  for (let i=0;i<365;i++) {
    const d = new Date();
    d.setDate(d.getDate() - i);
    if (dates.has(d.toDateString())) streak++;
    else break;
  }
  return streak;
}
