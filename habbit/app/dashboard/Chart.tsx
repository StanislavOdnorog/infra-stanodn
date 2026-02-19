'use client';
import React from 'react';
import ReactECharts from 'echarts-for-react';

export default function Chart({ data }: { data: { name: string; completed: number }[] }) {
  const option = {
    tooltip: {},
    xAxis: { type: 'category', data: data.map(d => d.name) },
    yAxis: { type: 'value' },
    series: [{ type: 'bar', data: data.map(d => d.completed) }]
  };
  return <ReactECharts option={option} style={{ height: 260 }} />;
}
