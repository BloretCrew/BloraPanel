#!/usr/bin/env python3
"""Collect a private endurance checkpoint into a credential-free report.

This only reads the owned run: it never logs in, signals, stops or resumes it.
--wait observes its existing worker and records the actual final result.
"""
import argparse
import datetime as dt
import hashlib
import json
import os
from pathlib import Path
import tempfile
import time

TERMINAL = {'PASSED_24H', 'PASSED_SHORT', 'FAILED', 'FAILED_CLEANUP', 'STOPPED'}


def read(path):
    if path.is_symlink() or not path.is_file():
        raise ValueError('Evidence must be a regular file')
    return json.loads(path.read_text(encoding='utf-8'))


def timestamp(value):
    parsed = dt.datetime.fromisoformat(value.replace('Z', '+00:00'))
    if parsed.tzinfo is None:
        raise ValueError('Evidence timestamps must include a timezone')
    return parsed


def evidence(root):
    config, state = read(root / 'run.json'), read(root / 'checkpoint.json')
    start = timestamp(state['segmentStartedAt'])
    finish = timestamp(state.get('finishedAt') or state['updatedAt'])
    events = []
    for path in sorted(root.glob('events.jsonl*')):
        if path.is_symlink() or not path.is_file():
            raise ValueError('Event evidence must be a regular file')
        try:
            lines = path.read_text(encoding='utf-8').splitlines()
        except FileNotFoundError:
            if state['status'] in TERMINAL:
                raise ValueError('Final event evidence disappeared')
            # A live logger can rotate between glob and read.
            continue
        for line in lines:
            try:
                item = json.loads(line)
            except json.JSONDecodeError:
                # The live worker may be in the middle of its final append.
                if state['status'] in TERMINAL:
                    raise ValueError('Final event evidence is incomplete')
                continue
            if start <= timestamp(item['utc']) <= finish:
                events.append(item)
    final = [e for e in events if e.get('stage') == 'final_checks_succeeded']
    cleanup = [e for e in events if e.get('stage') == 'cleanup_completed']
    final.sort(key=lambda e: timestamp(e['utc']))
    cleanup.sort(key=lambda e: timestamp(e['utc']))
    problems = []
    passed = state['status'] in {'PASSED_24H', 'PASSED_SHORT'}
    if passed:
        if state.get('error') or state.get('cleanupError'):
            problems.append('监督器记录了执行或清理错误')
        if not final or final[-1].get('sqliteIntegrity') != 'ok' or not final[-1].get('actualRestoreSnapshot'):
            problems.append('缺少本覆盖区间的实际恢复与SQLite完整性终态证据')
        if (not cleanup or cleanup[-1].get('servicesAlive') != [] or
                not final or timestamp(cleanup[-1]['utc']) < timestamp(final[-1]['utc']) or
                state.get('scheduleRemoved') is not True):
            problems.append('缺少所属计划及服务完成清理的证据')
        if (state.get('sampleCount', 0) < 121 or state.get('successfulSlots', 0) < 2 or
                len(state.get('phases', [])) != len(config['faultAfter'])):
            problems.append('采样、分钟槽或声明故障阶段未完整记录')
        if any(p.get('sameRunId') != state.get('runId') for p in state.get('phases', [])):
            problems.append('故障阶段没有保持原实例运行身份')
        for name, digest in config['binarySHA256'].items():
            binary = root / 'bin' / name
            if binary.is_symlink() or hashlib.sha256(binary.read_bytes()).hexdigest() != digest:
                problems.append('固定测试二进制与启动摘要不同')
        if state['status'] == 'PASSED_24H':
            if (config['seconds'] < 86400 or state.get('requiredSeconds') != config['seconds'] or
                    state.get('elapsedSeconds', 0) < 86400 or (finish - start).total_seconds() < 86400 or
                    len(set(state.get('datesUTC', []))) < 2):
                problems.append('24小时实际时长或跨UTC日期证据不足')
        elif config['seconds'] >= 86400:
            problems.append('长测仅获得短测状态，不能作为24小时通过')
    result = ('EVIDENCE_INCOMPLETE' if problems else state['status'])
    return config, state, result, final[-1] if final else None, problems


def render(root, config, state, result, final, problems, observer_note=''):
    is24 = result == 'PASSED_24H'
    meaning = ('本地24小时场景已通过，含最终真实恢复和所属资源清理。' if is24 else
               '仅本地短测通过，不能作为24小时通过。' if result == 'PASSED_SHORT' else
               '尚未完成或未通过，不能提升24小时验收状态。')
    rows = [
        '# 本地长测自动收取结果', '', f'状态：**{result}**。{meaning}', '',
        f'证据目录：`{root}`。报告仅提取白名单统计，不包含凭据、Cookie、服务命令或正文。', '',
        f'- 覆盖起点（UTC）：{timestamp(state["segmentStartedAt"]).isoformat()}',
        f'- 最后检查点（UTC）：{timestamp(state["updatedAt"]).isoformat()}',
        f'- 实际单调时间：{state.get("elapsedSeconds", 0):.3f}s；要求：{config["seconds"]}s',
        f'- 真实采样：{state.get("sampleCount", 0)}；成功分钟槽：{state.get("successfulSlots", 0)}',
        f'- UTC日期数：{len(set(state.get("datesUTC", [])))}；已完成故障阶段：{len(state.get("phases", []))}/{len(config["faultAfter"])}',
        f'- 最大普通采样间隔：{state.get("maxOrdinarySampleGapSeconds", 0):.3f}s', '',
        '| 资源 | RSS峰值（字节） | FD峰值 |', '| --- | ---: | ---: |'
    ]
    maxima = state.get('maxima', {})
    for name in ('master', 'daemon'):
        entry = maxima.get(name, {})
        rows.append(f'| {name} | {int(entry.get("rssBytes", 0))} | {int(entry.get("fds", 0))} |')
    rows += ['', f'归档峰值：{int(maxima.get("archiveBytes", 0))}B；日志helper RSS峰值：{int(maxima.get("logHelperRSSBytes", 0))}B；私有状态及证据峰值：{int(maxima.get("fixtureStateBytes", 0))}B。']
    if final and result in {'PASSED_24H', 'PASSED_SHORT'}:
        rows += ['', f'本覆盖区间最终事件确认SQLite integrity_check=ok、最新真实备份恢复正文匹配入账版本、所属定时任务唯一且成功；普通分钟槽：{int(final["ordinaryMinuteSlots"])}。所属计划已删除、服务清理列表为空。']
    if problems:
        rows += ['', '证据缺口：'] + [f'- {item}' for item in problems]
    if observer_note:
        rows += ['', observer_note]
    rows += ['', '本报告不覆盖当前E08多窗口性能、Windows真机、systemd、远端网络、设备掉电或操作系统通知中心；不得据此宣称全范围完成。', '']
    return '\n'.join(rows)


def write_report(path, text):
    if path.is_symlink() or not path.parent.is_dir():
        raise ValueError('Report destination must have a regular existing parent')
    descriptor, temporary = tempfile.mkstemp(prefix='.endurance-report-', dir=path.parent)
    try:
        with os.fdopen(descriptor, 'w', encoding='utf-8') as stream:
            os.fchmod(stream.fileno(), 0o644)
            stream.write(text)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, path)
    finally:
        Path(temporary).unlink(missing_ok=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', type=Path)
    parser.add_argument('--report', required=True, type=Path)
    parser.add_argument('--wait', action='store_true')
    parser.add_argument('--timeout-seconds', type=int, default=30000)
    args = parser.parse_args()
    if args.root.is_symlink():
        parser.error('Select the actual private endurance directory')
    root = args.root.resolve(strict=True)
    if (not root.name.startswith('endurance-') or root.stat().st_uid != os.getuid() or
            root.stat().st_mode & 0o077 or not 1 <= args.timeout_seconds <= 604800):
        parser.error('Select an owned private endurance-* directory and a bounded timeout')
    deadline = time.monotonic() + args.timeout_seconds
    last_write = 0
    while True:
        config, state, result, final, problems = evidence(root)
        terminal = state['status'] in TERMINAL
        note = ''
        stale = time.time() - (root / 'checkpoint.json').stat().st_mtime > 120
        expired = time.monotonic() >= deadline
        if args.wait and not terminal and (stale or expired):
            note = '收取器已结束等待：检查点超过120秒未更新或到达等待截止时间；没有停止、重启或修改测试服务。当前结果仍不能算通过。'
        if terminal or note or not args.wait or time.monotonic() - last_write >= 60:
            write_report(args.report, render(root, config, state, result, final, problems, note))
            last_write = time.monotonic()
        if terminal or note or not args.wait:
            print(json.dumps({'status': result, 'report': str(args.report.resolve()), 'terminal': terminal}, ensure_ascii=False))
            return 0 if result in {'PASSED_24H', 'PASSED_SHORT'} else 2
        time.sleep(5)


if __name__ == '__main__':
    raise SystemExit(main())
