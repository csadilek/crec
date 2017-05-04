package processor

import (
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

func removeNodes(context *Context, nodeNames []string) (*Context, error) {
	if !context.HTML {
		return context, nil
	}
	node, err := removeHTMLNodes(context.Content.(*html.Node), nodeNames)
	if err != nil {
		return nil, err
	}
	return &Context{Content: node, HTML: true, Result: context.Result}, nil
}

func removeHTMLNodes(node *html.Node, nodeNames []string) (*html.Node, error) {
	removeMatchingNodes(node, nodeNames)
	return node, nil
}

func removeMatchingNodes(n *html.Node, nodeNames []string) {
	nodes := findNodesToRemove(n, nodeNames)
	for _, a := range nodes {
		n.RemoveChild(a)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		removeMatchingNodes(c, nodeNames)
	}
}

func findNodesToRemove(n *html.Node, nodeNames []string) []*html.Node {
	nodes := make([]*html.Node, 0)

	nodeMap := make(map[string]bool)
	for _, node := range nodeNames {
		nodeMap[node] = true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && nodeMap[c.Data] == true {
			nodes = append(nodes, c)
		}
	}

	return nodes
}

func findAttributeValueOfFirstMatch(context *Context, elem string, attr string) (string, error) {
	if !context.HTML {
		return "", nil
	}
	node := context.Content.(*html.Node)
	img := cascadia.MustCompile(elem).MatchFirst(node)
	if img != nil {
		for _, a := range img.Attr {
			if a.Key == attr {
				if a.Val != "" {
					return a.Val, nil
				}
			}
		}
	}
	return "", nil
}
