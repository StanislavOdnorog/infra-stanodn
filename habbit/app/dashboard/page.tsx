import { getCurrentUser } from '../../lib/auth';
import { redirect } from 'next/navigation';
import { prisma } from '../../lib/db';
import { isHabitDueToday, calcStreak } from '../../lib/habits';
import Chart from './Chart';

export default async function Dashboard() {
  const user = await getCurrentUser();
  if (!user) redirect('/');
  const habits = await prisma.habit.findMany({ where: { userId: user.id, status: 'ACTIVE' }, include: { logs: true } });
  const today = habits.filter(h => isHabitDueToday(h as any, h.logs as any));
  const chartData = habits.map(h => ({ name: h.title, completed: h.logs.length }));

  return (
    <div className="grid">
      <div className="row" style={{ justifyContent: 'space-between' }}>
        <h1>Today</h1>
        <form action="/api/auth/logout" method="post">
          <button type="submit">Logout</button>
        </form>
      </div>

      <div className="card grid">
        <h3>Create habit</h3>
        <form action="/api/habits" method="post" className="row">
          <input name="title" placeholder="Habit title" />
          <button type="submit">Add</button>
        </form>
      </div>

      {today.length === 0 && <div className="card">No habits scheduled today.</div>}
      {today.map(h => (
        <div className="card" key={h.id}>
          <div className="row">
            <strong>{h.title}</strong>
            <span>Streak: {calcStreak(h.logs as any)}</span>
          </div>
          <form action={`/api/habits/${h.id}/toggle`} method="post">
            <button type="submit">Mark done</button>
          </form>
        </div>
      ))}

      <div className="card">
        <h3>Completion chart</h3>
        <Chart data={chartData} />
      </div>
    </div>
  );
}
