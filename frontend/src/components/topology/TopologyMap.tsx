import React, { useEffect, useRef } from 'react';
import { Network } from 'vis-network';
import { DataSet } from 'vis-data';

export interface TopologyNode {
  id: string;
  label: string;
  title: string;
  group: string;
  color: string;
  shape: string;
  size: number;
}

export interface TopologyEdge {
  from: string;
  to: string;
  label: string;
  dashes: boolean;
  color: string;
  width: number;
}

interface TopologyMapProps {
  nodes: TopologyNode[];
  edges: TopologyEdge[];
  onNodeClick?: (nodeId: string) => void;
  darkMode?: boolean;
}

const TopologyMap: React.FC<TopologyMapProps> = ({ nodes, edges, onNodeClick, darkMode = true }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const networkRef = useRef<Network | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    const nodeData = new DataSet<TopologyNode>(nodes);
    const edgeData = new DataSet<TopologyEdge>(edges);

    const options = {
      autoResize: true,
      physics: {
        enabled: true,
        solver: 'forceAtlas2Based' as const,
        forceAtlas2Based: {
          gravitationalConstant: -50,
          centralGravity: 0.01,
          springLength: 200,
          springConstant: 0.08,
        },
        stabilization: {
          iterations: 200,
          updateInterval: 25,
        },
      },
      interaction: {
        dragNodes: true,
        dragView: true,
        zoomView: true,
        hover: true,
        tooltipDelay: 300,
      },
      nodes: {
        font: {
          color: darkMode ? '#d1d5db' : '#374151',
          size: 12,
          face: 'Inter, sans-serif',
        },
        borderWidth: 2,
        borderWidthSelected: 3,
      },
      edges: {
        smooth: {
          type: 'continuous' as const,
        },
        font: {
          size: 10,
          color: '#6b7280',
        },
      },
      groups: {
        network: { color: { background: '#1677ff', border: '#0958d9' } },
        server: { color: { background: '#fa8c16', border: '#d46b08' } },
        endpoint: { color: { background: '#13c2c2', border: '#08979c' } },
        unknown: { color: { background: '#8c8c8c', border: '#595959' } },
      },
    };

    const network = new Network(containerRef.current, { nodes: nodeData, edges: edgeData }, options);
    networkRef.current = network;

    if (onNodeClick) {
      network.on('click', (params: any) => {
        if (params.nodes.length > 0) {
          onNodeClick(params.nodes[0]);
        }
      });
    }

    network.once('stabilizationIterationsDone', () => {
      network.fit({ animation: { duration: 1000 } });
    });

    return () => {
      network.destroy();
    };
  }, []);

  useEffect(() => {
    if (!networkRef.current) return;
    const nodeData = new DataSet<TopologyNode>(nodes);
    const edgeData = new DataSet<TopologyEdge>(edges);
    networkRef.current.setData({ nodes: nodeData, edges: edgeData });
  }, [nodes, edges]);

  return (
    <div
      ref={containerRef}
      className="w-full h-full"
      style={{ background: darkMode ? '#0d1117' : '#f5f5f5' }}
    />
  );
};

export default TopologyMap;
