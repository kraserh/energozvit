#!/bin/bash

WORKDIR=Output
JOBNAME=report

mkdir -p $WORKDIR
cat $JOBNAME.tex.tmpl | energozvit-tmpl "$1" $2 > $WORKDIR/$JOBNAME.tex
cd $WORKDIR
pdflatex -interaction=nonstopmode $JOBNAME.tex \
	&& xdg-open $JOBNAME.pdf &> /dev/null
